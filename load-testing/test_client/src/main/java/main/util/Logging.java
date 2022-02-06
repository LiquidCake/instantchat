package main.util;

import org.apache.commons.lang3.exception.ExceptionUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.FileWriter;
import java.io.PrintWriter;
import java.text.SimpleDateFormat;
import java.util.Calendar;
import java.util.Date;

public class Logging {
    private static final Logger LOGGER = LoggerFactory.getLogger(Util.class);

    public static final String LOG_DATE_FORMAT_NOW = "yyyy-MM-dd HH:mm:ss";
    public static final SimpleDateFormat logDateFormat = new SimpleDateFormat(LOG_DATE_FORMAT_NOW);

    public static boolean debugLogEnabled = false;

    public static String getLogDateString(final long timestamp) {
        return logDateFormat.format(new Date(timestamp));
    }

    public static void logInfo(final String roomName, final String logLine) {
        logLine("INFO", roomName, logLine);
    }

    public static void logError(final String roomName, final String logLine) {
        logLine("ERROR", roomName, logLine);
    }

    public static void logError(final String roomName, final String logLine, final Throwable t) {
        logLine("ERROR", roomName, logLine + ". Exception: " + ExceptionUtils.getStackTrace(t));
    }

    public static void logTrace(final String roomName, final String logLine) {
        if (debugLogEnabled) {
            logLine("TRACE", roomName, logLine);
        }
    }

    private static void logLine(final String level, final String logFileNamePart, final String logLine) {
        final String logFileName = "./logs/log-" + logFileNamePart + ".log";

        try (final PrintWriter printWriter = new PrintWriter(new FileWriter(logFileName, true))) {
            printWriter.println(String.format("%s [%s] %s", getLogDateNowString(), level, logLine));
        } catch (final Exception e) {
            LOGGER.error("failed to open file to log: " + logFileName);
        }
    }

    private static String getLogDateNowString() {
        return logDateFormat.format(Calendar.getInstance().getTime());
    }


    public static void logHttpInfo(final String logLine) {
        logLine("INFO", "http", logLine);
    }

    public static void logHttpError(final String logLine) {
        logLine("ERROR", "http", logLine);
    }

    public static void logHttpError(final String logLine, final Throwable t) {
        logLine("ERROR", "http", logLine + ". Exception: " + ExceptionUtils.getStackTrace(t));
    }
}
