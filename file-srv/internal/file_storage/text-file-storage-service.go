package file_storage

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"instantchat.rooms/instantchat/file-srv/internal/util"
)

type TextFileContentsCacheItem struct {
	fileLock           sync.Mutex
	lastCheckTimestamp int64
	fileTextContent    []byte
}

/* Constants */

const UploadTextFilesDirPath = "/var/file-srv/uploaded-files/text/"

const ClearOldCacheItemsFuncDelay = 30 * time.Second
const TextFilesCacheTTL = 30 * time.Second

const DeleteOldTextFilesFuncDelay = 1 * time.Hour
const TextFileGroupUnchangedTTL = 6 * time.Hour

/* Variables */

var dirCreationMutex = sync.Mutex{}

var textFilesCache = map[string]*TextFileContentsCacheItem{}
var textFilesCacheMutex = sync.Mutex{}

func SaveTextFileToDisk(fileContent string, fileName string, fileGroupPrefix string) error {
	if strings.Contains(fileName, "..") {
		return errors.New("bad file name")
	}

	//create dir to store file if not exists
	pathToFileGroup := filepath.Join(UploadTextFilesDirPath, fileGroupPrefix)

	dirCreationMutex.Lock()
	err := os.MkdirAll(pathToFileGroup, os.ModePerm)
	dirCreationMutex.Unlock()

	if err != nil {
		return err
	}

	//created with perm. 0666 by default
	file, err := os.Create(filepath.Join(pathToFileGroup, fileName))

	if err != nil {
		return err
	}

	defer file.Close()

	bytesWritten, err := file.WriteString(fileContent)

	if err != nil {
		return err
	}

	if bytesWritten == 0 {
		return errors.New("0 bytes written to disk")
	}

	return nil
}

func ReadTextFileFromDisk(fileName string, fileGroupPrefix string) ([]byte, error) {
	if strings.Contains(fileName, "..") {
		return nil, errors.New("bad file name")
	}

	pathToFile := filepath.Join(UploadTextFilesDirPath, fileGroupPrefix, fileName)

	textFilesCacheMutex.Lock()

	_, fileAlreadyExistsInCache := textFilesCache[pathToFile]

	if !fileAlreadyExistsInCache {
		textFilesCache[pathToFile] = &TextFileContentsCacheItem{
			fileLock:           sync.Mutex{},
			lastCheckTimestamp: 0,
			fileTextContent:    nil,
		}
	}

	textFilesCacheMutex.Unlock()

	textFileCacheItem, _ := textFilesCache[pathToFile]

	//lock current file's cache item to prevent concurrent queries to same file
	textFileCacheItem.fileLock.Lock()
	defer textFileCacheItem.fileLock.Unlock()

	if textFileCacheItem.fileTextContent == nil {
		bytes, err := ioutil.ReadFile(pathToFile)

		if err != nil {
			return nil, err
		}

		textFileCacheItem.fileTextContent = bytes
		textFileCacheItem.lastCheckTimestamp = time.Now().UnixNano()
	}

	return textFileCacheItem.fileTextContent, nil
}

func StartClearOldCacheItemsFuncPeriodical() {
	ticker := time.NewTicker(ClearOldCacheItemsFuncDelay)

	for {
		select {
		case <-ticker.C:
			clearOldCacheItems()
		}
	}
}

func StartDeleteOldTextFilesFuncPeriodical() {
	ticker := time.NewTicker(DeleteOldTextFilesFuncDelay)

	for {
		select {
		case <-ticker.C:
			deleteOldTextFiles()
		}
	}
}

func clearOldCacheItems() {
	var entriesToRemoveFromCacheArr []string

	textFilesCacheMutex.Lock()

	for filePath, info := range textFilesCache {
		if info.lastCheckTimestamp > 0 &&
			time.Now().UnixNano()-info.lastCheckTimestamp > TextFilesCacheTTL.Nanoseconds() {

			entriesToRemoveFromCacheArr = append(entriesToRemoveFromCacheArr, filePath)
		}
	}

	for _, filePath := range entriesToRemoveFromCacheArr {
		delete(textFilesCache, filePath)
	}

	textFilesCacheMutex.Unlock()
}

func deleteOldTextFiles() {
	timeNow := time.Now()

	textFilesGroupDir, err := ioutil.ReadDir(UploadTextFilesDirPath)

	if err != nil {
		util.LogSevere("Failed to list contents of text file uploads DIR ('%s'): '%s'", UploadTextFilesDirPath, err)

		return
	}

	//iterate all text files group folders, if for some folder no modifications happened for a long time (last file was added too long ago) - delete that folder
	for _, childDir := range textFilesGroupDir {
		//when the file was created inside dir for a last time
		dirModificationTime := childDir.ModTime()

		if dirModificationTime.Add(TextFileGroupUnchangedTTL).Before(timeNow) {
			textFilesGroupDirPath := filepath.Join(UploadTextFilesDirPath, childDir.Name())
			err := os.RemoveAll(textFilesGroupDirPath)

			if err != nil {
				util.LogSevere("Failed to delete text files group DIR ('%s'): '%s'", textFilesGroupDirPath, err)
			}
		}
	}
}
