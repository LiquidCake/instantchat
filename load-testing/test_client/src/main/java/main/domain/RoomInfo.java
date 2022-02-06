package main.domain;

import com.fasterxml.jackson.annotation.JsonProperty;

public class RoomInfo {
    @JsonProperty("n")
    private String name;
    @JsonProperty("p")
    private String password;

    public RoomInfo() {}

    public RoomInfo(final String name, final String password) {
        this.name = name;
        this.password = password;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }

    @Override
    public String toString() {
        return "RoomInfo{" +
                "name='" + name + '\'' +
                ", password='" + password + '\'' +
                '}';
    }
}
