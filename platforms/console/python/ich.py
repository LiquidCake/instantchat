#!/usr/bin/env python3

#
# askii rendering is skipped - if you want to enable - include your own code to render ascii art from image url
# e.g. asciify.py (https://github.com/RameshAditya/asciify) can be copied right into this script
#
def get_image_ascii():
    print_green('| cant render ascii for image (no renderer supplied)')

### version - 1.2

import websocket
import rel
from threading import *

import http.cookiejar
from urllib.request import urlopen
import urllib.parse
import urllib
import json
import re
import time
import sys
from sys import platform
import os

os.system("")  #enables ansi escape characters (coloring) in terminal for Windows 10+

try:
    import readline #importing this module allows keyboard arrows in terminal input on linux (needs tty fix before exit though)
except ImportError:
    pass


#--dev: change to your address
SERVER_ADDR = 'https://myinstantchat.org'
ORIGIN = SERVER_ADDR

PICK_BACKEND_ENDPOINT = '/pick_backend'
GET_TEXT_FILE_ENDPOINT = '/get_text_file'

ROOM_CREDS_MIN_LENGTH = 3
ROOM_CREDS_MAX_LENGTH = 100

MAX_TEXT_MESSAGE_LENGTH = 10000

MESSAGE_META_MARKER_TYPE_DRAWING = '$#$meta_marker_is_drawing$#$'

COMMANDS = {
    'RoomCreateJoin': 'R_C_J',
    'RoomCreateJoinAuthorize': 'R_C_J_AU',
    'RoomCreate': 'R_C',
    'RoomJoin': 'R_J',
    'RoomChangeUserName': 'R_CH_UN',
    'RoomChangeDescription': 'R_CH_D',
    'RoomMembersChanged': 'R_M_CH',
	
    'TextMessage': 'TM',
    'TextMessageEdit': 'TM_E',
    'TextMessageDelete': 'TM_D',
    'TextMessageSupportOrReject': 'TM_S_R',
    'AllTextMessages': 'ALL_TM',
	
    'UserDrawingMessage': 'DM',
	
    'Error': 'ER',
    'RequestProcessed': 'RP',
	
    'NotifyMessagesLimitApproaching': 'N_M_LIMIT_A',
    'NotifyMessagesLimitReached': 'N_M_LIMIT_R',
}

BUSINESS_ERRORS = {
    101: {'name': "WsServerError",                             'code': 101, 'text': "server error"},
    102: {'name': "WsConnectionError",                         'code': 102, 'text': "connection error"},
    103: {'name': "WsInvalidInput",                            'code': 103, 'text': "invalid input"},

    201: {'name': "WsRoomExists",                              'code': 201, 'text': "room with this name already exists"},
    202: {'name': "WsRoomNotFound",                            'code': 202, 'text': "room not found"},
    203: {'name': "WsRoomInvalidPassword",                     'code': 203, 'text': "invalid room password"},
    204: {'name': "WsRoomUserNameTaken",                       'code': 204, 'text': "provided user name is already taken"},
    205: {'name': "WsRoomUserNameValidationError",             'code': 205, 'text': "invalid room user name length"},
    206: {'name': "WsRoomNotAuthorized",                       'code': 206, 'text': "not authorized to join this room"},
    207: {'name': "WsRoomMessageTooLargeError",                'code': 207, 'text': "message is too long"},
    208: {'name': "WsRoomIsFullError",                         'code': 208, 'text': "room is full"},
    209: {'name': "WsRoomUserDuplication",                     'code': 209, 'text': "user connected to this room from another client"},

    301: {'name': "WsRoomCredsValidationErrorBadLength",       'code': 301, 'text': "invalid room name length"},
    302: {'name': "WsRoomCredsValidationErrorNameForbidden",   'code': 302, 'text': "room name is forbidden"},
    303: {'name': "WsRoomCredsValidationErrorNameHasBadChars", 'code': 303, 'text': "room name contains bad characters"},
    304: {'name': "WsRoomValidationErrorBadDescriptionLength", 'code': 304, 'text': "invalid room description length"},
}

SERVER_STATUS_ONLINE =        'online'
SERVER_STATUS_SHUTTING_DOWN = 'shutting_down'
SERVER_STATUS_RESTARTING =    'restarting'

lock = Lock()

ws = None

ws_connected = False
shutdown_requested = False
logged_into_room = False

room_name = None
room_passwd = None
send_stdin_and_exit = False
no_color = False
no_askii = False

session_cookie = None

last_user_list_timestamp = None
last_user_info_by_id_cache = {}

version = None

def on_message(ws, msg):
    global room_name
    global logged_into_room
    global version
    global last_user_list_timestamp
    global last_user_info_by_id_cache
    
    message_frame = json.loads(msg)

    match message_frame['c']:
        case 'ER':
            type_code = int(message_frame['m'][0]['t'])
            business_error = BUSINESS_ERRORS[type_code]
            error_text = business_error['text']

            print_red('| error: {}'.format(error_text))

            if type_code in [201, 203, 208, 301, 302, 303]:
                shutdown()
                return
            
            if type_code in [207]:
                return

        case 'RP':
            if 'rq' in message_frame and message_frame['rq'] == 'room_c_j_done':
                print_green('| system: logged into room "{}" as "{}"'
                            .format(room_name, get_user_name_by_id(message_frame['uId'])))

                logged_into_room = True

                version = message_frame['bN']

                print_green('| version: {}'.format(version))

        case 'R_M_CH':
            new_user_list_timestamp = message_frame['cAt']

            if last_user_list_timestamp == None or new_user_list_timestamp > last_user_list_timestamp:
                last_user_list_timestamp = new_user_list_timestamp

                new_user_info_by_id_cache = {}
                new_users_list = message_frame['rU']
                
                for user in new_users_list:
                    new_user_info_by_id_cache[user['uId']] = {
                        'id': user['uId'], 
                        'name': user['n'], 
                        'isAnon': user['an'], 
                        'isOnlineInRoom': user['o']
                        }
                
                for user_id in new_user_info_by_id_cache:
                    user_info = new_user_info_by_id_cache[user_id]

                    if user_info['isOnlineInRoom'] and user_id not in last_user_info_by_id_cache:
                        user_name = urllib.parse.unquote(user_info['name'])
                        print_green('| user joined: "{}"'.format(user_name))
                    
                    elif user_info['isOnlineInRoom'] and user_info['name'] != last_user_info_by_id_cache[user_id]['name']:
                        new_user_name = urllib.parse.unquote(user_info['name'])
                        old_user_name = urllib.parse.unquote(last_user_info_by_id_cache[user_id]['name'])
                        print_green('| user "{}" changed name to "{}"'.format(old_user_name, new_user_name))
                
                last_user_info_by_id_cache = new_user_info_by_id_cache
            
        case 'ALL_TM':
            if not send_stdin_and_exit:
                for message in message_frame['m']:
                    print_text_message(message)

        case 'TM':
            if not send_stdin_and_exit:
                message = message_frame['m'][0]
                print_text_message(message)

        case 'TM_E':
            message = message_frame['m'][0]
            text_message = urllib.parse.unquote(message['t'])
            print_green('| message edited: #{}: "{}"'.format(message['id'], text_message))

        case 'TM_D':
            print_green('| message deleted: #{}'.format(message_frame['m'][0]['id']))

        case 'DM':
            if not send_stdin_and_exit:
                message = message_frame['m'][0]
                print_drawing_message(message)
        
        case 'TM_S_R':
            message = message_frame['m'][0]
            message_id = message['id']
            print_green('| message #{} voted: +{} / -{}'.format(message_id, message['sC'], message['rC']))

        case 'R_CH_D':
            message = message_frame['m'][0]
            server_status = message_frame.get('sS', None)

            if server_status == SERVER_STATUS_SHUTTING_DOWN:
                print_red('| system: server is shutting down for maintenance in a minute, please save your data')
            elif server_status == SERVER_STATUS_RESTARTING:
                print_red('| system: server is restarting in a minute, please save your data')

            new_room_description = urllib.parse.unquote(message['t'])
            
            if new_room_description.strip():
                print_green('| new room description: {}'.format(new_room_description))

        case 'N_M_LIMIT_A':
            print_red('| system: {}'.format('room is approaching messages limit, old messages will be removed soon'))

        case 'N_M_LIMIT_R':
            print_red('| system: {}'.format('room messages limit reached, old messages were removed'))

        case _:
            print_red('| system: unknown message type: {}'.format(message_frame['c']))

def on_error(ws, error):
    print_red('| error (connection): {}'.format(str(error)))

def on_close(ws, close_status_code, close_msg):
    print_green('| system: connection closed')

def on_open(ws):
    global ws_connected
    ws_connected = True

    print_green('| system: connected')

    log_into_room()

#run in main thread
def socket_loop(ws_endpoint_addr, session_cookie):
    global ws
    global shutdown_requested

    #websocket.enableTrace(True)
    ws = websocket.WebSocketApp(ws_endpoint_addr,
                            on_open=on_open,
                            on_message=on_message,
                            on_error=on_error,
                            on_close=on_close,
                            cookie='session={}'.format(session_cookie))

    while not shutdown_requested:
        try:
            disconnect_ws()

            print_green('| system: connecting ...')
            time.sleep(1)

            ws.run_forever(dispatcher=rel, origin=ORIGIN, ping_timeout=5, ping_interval=10, reconnect=5)
            #sig interrupt handler
            rel.signal(2, shutdown)
            rel.dispatch()
        except Exception as e:
            print_red('| system: connection lost: {}'.format(e))

    print_green('| system: exiting')

def shutdown():
    global shutdown_requested
    shutdown_requested = True

    rel.abort()

    #fix tty on linux
    if platform != 'win32':
        os.system('stty sane')

#call from main thread
def disconnect_ws():
    global ws_connected
    global logged_into_room
    global ws

    try:
        lock.acquire()

        rel.abort()
        ws.close()

        ws_connected = False
        logged_into_room = False

        print_green('| system: disconnected')

    finally:
        lock.release()

def send_msg(msg_dict):
    global ws_connected

    if ws_connected:
        try:
            lock.acquire()
            ws.send(json.dumps(msg_dict))
        except:
            print_red('| error: failed to send message')
        finally:
            lock.release()

def keep_alive_handler():
    global shutdown_requested

    while not shutdown_requested:
        send_keep_alive()
        time.sleep(5)

def send_keep_alive():
    send_msg({'kA': 'OK'})

def is_chat_command(command):
    if '/help' == command\
        or re.match('^/del \d+$', command)\
        or '/who' == command \
        or re.match('^/vote\+ \d+$', command)\
        or re.match('^/vote\- \d+$', command)\
        or re.match('^/changename \w+$', command)\
        or '/getsession' == command:
        return True

    return False

def execute_chat_command(command):
    if command == '/help':
        print('Chat commands:')
        print('\t/help               - show this list')
        print('\t/del 123            - delete message with id #123')
        print('\t/who                - print list of room users')
        print('\t/vote+ 123          - support message with id #123')
        print('\t/vote- 123          - reject message with id #123')
        print('\t/changename newname - change your name in this room')
        print('\t/getsession         - print your session token to be re-used later (see env vars)')

    elif re.match('^/del \d+$', command):        
        send_msg({
            'c': COMMANDS['TextMessageDelete'],
            'r': {
                'n': room_name
            },
            'm': {
                'id': int(command.replace('/del ', ''))
            }
        })
    
    elif '/who' == command: 
        print_green('Users online:')

        for user_id in last_user_info_by_id_cache:
            user_info = last_user_info_by_id_cache[user_id]

            if user_info['isOnlineInRoom']:
                print_green('| {}'.format(user_info['name']))

    elif re.match('^/vote\+ \d+$', command):
           send_msg({
            'c': COMMANDS['TextMessageSupportOrReject'],
            'r': {
                'n': room_name
            },
            'srM': True,
            'm': {
                'id': int(command.replace('/vote+ ', ''))
            }
        })
           
    elif re.match('^/vote\- \d+$', command):
             send_msg({
            'c': COMMANDS['TextMessageSupportOrReject'],
            'r': {
                'n': room_name
            },
            'srM': False,
            'm': {
                'id': int(command.replace('/vote- ', ''))
            }
        })
             
    elif re.match('^/changename \w+$', command):
           send_msg({
            'c': COMMANDS['RoomChangeUserName'],
            'r': {
                'n': room_name
            },
            'uN': urllib.parse.quote(command.replace('/changename ', ''))
        })
           
    elif '/getsession' == command:
        print_green('| your long-term session token: {}'.format(session_cookie))

def user_input_handler(*args):
    global logged_into_room
    global room_name
    global shutdown_requested
    global send_stdin_and_exit

    while not shutdown_requested:
        if send_stdin_and_exit:
            user_msg_text = ''.join(sys.stdin.readlines())
        else:
            try:
                user_msg_text = input()
            except:
                continue
            
            user_msg_text = user_msg_text.strip().replace('\r', '').replace('\n', '')

            if is_chat_command(user_msg_text):
                execute_chat_command(user_msg_text)
                
                continue

        if len(user_msg_text.strip()) == 0:
            continue

        if len(user_msg_text) >= MAX_TEXT_MESSAGE_LENGTH:
            print_red('| error: message is too long')
            
            if send_stdin_and_exit:
                shutdown()
            
            continue

        while not logged_into_room:
            time.sleep(1)
        
        send_msg({
            'c': COMMANDS['TextMessage'],
            'r': {
                'n': room_name
            },
            'm': {
                't': urllib.parse.quote(user_msg_text)
            }
        })

        if send_stdin_and_exit:
            print_green('| stdin message sent')

            shutdown()

def log_into_room():
    send_msg({'p': 'cp'})
    send_msg({
                'c': COMMANDS['RoomCreateJoin'],
                'uN': None,
                'rq': 'room_c_j_done',
                'r': {
                    'n': room_name,
                    'p': room_passwd
                }
            })

def print_text_message(message):
    global last_user_info_by_id_cache

    message_id = message['id']
    text_message = urllib.parse.unquote(message['t'])

    if MESSAGE_META_MARKER_TYPE_DRAWING in text_message:
        print_drawing_message(message)
        return

    author_user_id = message['uId']
    author_user_name = get_user_name_by_id(author_user_id)

    print('{}: #{} {}'.format(author_user_name, message_id, text_message))

def print_drawing_message(message):
    global no_askii

    author_user_id = message['uId']
    author_user_name = get_user_name_by_id(author_user_id)

    if no_askii:
        print_green('| drawing message from user {} skipped (-noaskii set)'.format(author_user_name))
        
        return

    message_text = urllib.parse.unquote(message['t'])
    text_meta = message_text.split(MESSAGE_META_MARKER_TYPE_DRAWING)[1];
    file_name = text_meta.split("@")[0];
    file_group_name = text_meta.split("@")[1];

    url = '{}{}?file_name={}&file_group_prefix={}'.format(SERVER_ADDR, GET_TEXT_FILE_ENDPOINT, file_name, file_group_name)

    print_green('| drawing by {}:'.format(author_user_name))
    print(get_image_ascii(url))
    print('\n')

def print_green(msg):
    global no_color
    if no_color:
        print(msg)
    else:
        print('\033[92m{}\033[00m'.format(msg))

def print_red(msg):
    global no_color
    if no_color:
        print(msg)
    else:
        print('\033[91m{}\033[00m'.format(msg))

def get_user_name_by_id(user_id):
    return urllib.parse.unquote(last_user_info_by_id_cache[user_id]['name'])

def get_backend_url():
    url = '{}{}?roomName={}'.format(SERVER_ADDR, PICK_BACKEND_ENDPOINT, urllib.parse.quote(room_name))

    try:
        with urlopen(url, timeout=5) as response:
            body = json.loads(response.read().decode())

        error = body.get('e', None)
        if error:
            print_red('| error: {}'.format(error))
            exit(1)
        
        return body['bA']
    except Exception as e:
        print_red('| error: cannot get backend address - {}'.format(e))

        exit(1)
    
def get_session_cookie():
    session_token = os.getenv('INSTANTCHAT_SESSION')
    if session_token != None:
        return session_token

    try:
        cookiejar = http.cookiejar.CookieJar()
        cookieproc = urllib.request.HTTPCookieProcessor(cookiejar)
        opener = urllib.request.build_opener(cookieproc)

        with opener.open(SERVER_ADDR, timeout=5) as response:
            for cookie in cookiejar:
                if cookie.name == 'session':
                    return cookie.value

    except Exception as e:
        print_red('| error: cannot create session - {}'.format(e))

        exit(1)

if __name__ == '__main__':
    args_count = len(sys.argv)

    if args_count < 2 or sys.argv[1] in ['help', '-help', '--help', '/?', '?']:
        print('usage: {} room_name [-p my_password -stdin -nocolor -noascii]\n'.format(sys.argv[0]))
        print('optional params:')
        print('\t-p my_password - password for room (if any)')
        print('\t-stdin         - sends input that was piped to stdin and exits')
        print('\t-nocolor       - dont use terminal coloring')
        print('\t-noaskii       - dont show user drawings as ascii pictures')
        print('')
        print('Chat commands:')
        print('\t/help               - show this list')
        print('\t/del 123            - delete message with id #123')
        print('\t/who                - print list of room users')
        print('\t/vote+ 123          - support message with id #123')
        print('\t/vote- 123          - reject message with id #123')
        print('\t/changename newname - change your name in this room')
        print('\t/getsession         - print your session token to be re-used later (see env vars)')
        print('')
        print('Environment variables:')
        print('INSTANTCHAT_SESSION=eyJzZXNzaW9uVVVJRCI... - optional session token (see /getsession command)')
        print('')

        exit(0)

    room_name = sys.argv[1]

    for i in range(0, args_count):
        arg = sys.argv[i]

        if arg == '-p':
            room_passwd = sys.argv[i + 1]
            
        if arg == '-stdin':
            send_stdin_and_exit = True

        if arg == '-nocolor':
            no_color = True

        if arg == '-noaskii':
            no_askii = True
    
    session_cookie = get_session_cookie()
    ws_endpoint_addr = 'wss://{}/ws_entry'.format(get_backend_url())
   
    user_input_thread = Thread(target=user_input_handler, daemon=True)
    user_input_thread.start()

    keep_alive_thread = Thread(target=keep_alive_handler, daemon=True)
    keep_alive_thread.start()
    
    socket_loop(ws_endpoint_addr, session_cookie)
