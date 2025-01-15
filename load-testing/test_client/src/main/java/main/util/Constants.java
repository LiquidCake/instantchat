package main.util;

import com.fasterxml.jackson.databind.ObjectMapper;

public class Constants {
    public static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    //--dev: change to your address
    //aux-srv address
    public static final String SERVER_ROOT_ADDR = "myinstantchat.org";

    public static final String HTTP_PROTOCOL = "https://";
    public static final String WS_PROTOCOL = "wss://";

    public static final String AUX_SRV_PICK_BACKEND_ENDPOINT = HTTP_PROTOCOL + SERVER_ROOT_ADDR + "/pick_backend?roomName=";

    public static final String WS_ENDPOINT = "/ws_entry";

    public static final String ROOM_PASSWORD = "123qwe_SOME$!%";
    public static final String USER_SESSION_COOKIE_NAME = "session";

    public static final int SOCKET_TIMEOUT_MS = 5000;

    public static final int SEND_MESSAGE_DELAY_MS = 3000;

    public static final String MESSAGE_TEXT_UNIQUE_PART_SPLITTER = "-!unique-part-ends!-";
}
