package main.domain;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Arrays;

public class PickBackendResponse {
    @JsonProperty("bA")
    private String backendInstanceAddr;
    @JsonProperty("e")
    private String errorMessage;
    @JsonProperty("aN")
    private String[] AlternativeRoomNamePostfixes;

    public PickBackendResponse() {}

    public String getBackendInstanceAddr() {
        return backendInstanceAddr;
    }

    public void setBackendInstanceAddr(String backendInstanceAddr) {
        this.backendInstanceAddr = backendInstanceAddr;
    }

    public String getErrorMessage() {
        return errorMessage;
    }

    public void setErrorMessage(String errorMessage) {
        this.errorMessage = errorMessage;
    }

    public String[] getAlternativeRoomNamePostfixes() {
        return AlternativeRoomNamePostfixes;
    }

    public void setAlternativeRoomNamePostfixes(String[] alternativeRoomNamePostfixes) {
        AlternativeRoomNamePostfixes = alternativeRoomNamePostfixes;
    }

    @Override
    public String toString() {
        return "PickBackendResponse{" +
                "backendInstanceAddr='" + backendInstanceAddr + '\'' +
                ", errorMessage='" + errorMessage + '\'' +
                ", AlternativeRoomNamePostfixes=" + Arrays.toString(AlternativeRoomNamePostfixes) +
                '}';
    }
}
