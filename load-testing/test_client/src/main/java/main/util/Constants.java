package main.util;

import com.fasterxml.jackson.databind.ObjectMapper;

public class Constants {
    public static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    //aux-srv address
    public static final String SERVER_ROOT_ADDR = "192.168.1.100";

    public static final String HTTP_PROTOCOL = "http://";
    public static final String WS_PROTOCOL = "ws://";

    public static final String AUX_SRV_PICK_BACKEND_ENDPOINT = HTTP_PROTOCOL + SERVER_ROOT_ADDR + "/pick_backend?roomName=";

    public static final String WS_ENDPOINT = WS_PROTOCOL + SERVER_ROOT_ADDR + "/ws_entry?backendHost=";

    public static final String ROOM_PASSWORD = "123qwe_SOME$!%";
    public static final String USER_SESSION_COOKIE_NAME = "session";

    public static final int SOCKET_TIMEOUT_MS = 5000;

    public static final int SEND_MESSAGE_DELAY_MS = 3000;

    public static final String MESSAGE_TEXT_UNIQUE_PART_SPLITTER = "-!unique-part-ends!-";
}
