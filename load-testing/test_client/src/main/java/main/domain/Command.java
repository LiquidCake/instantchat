package main.domain;

public class Command {
    public static String RoomCreateJoin =                  "R_C_J";
    public static String RoomCreate =                      "R_C";
    public static String RoomJoin =                        "R_J";
    public static String RoomChangeUserName =              "R_CH_UN";
    public static String RoomChangeUserDescription =       "R_CH_D";
    public static String RoomMembersChanged =              "R_M_CH";

    public static String TextMessage =                 "TM";
    public static String TextMessageEdit =             "TM_E";
    public static String TextMessageDelete =           "TM_D";
    public static String TextMessageSupportOrReject =  "TM_S_R";
    public static String AllTextMessages =             "ALL_TM";

    public static String Error =                           "ER";
    public static String RequestProcessed =                "RP";

    public static String NotifyMessagesLimitApproaching =  "N_M_LIMIT_A";
    public static String NotifyMessagesLimitReached =      "N_M_LIMIT_R";
}
