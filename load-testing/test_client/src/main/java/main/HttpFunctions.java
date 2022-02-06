package main;

import main.domain.PickBackendResponse;
import main.util.Constants;
import main.util.Logging;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.time.Instant;
import java.util.List;
import java.util.Map;

public class HttpFunctions {
    private static final Logger LOGGER = LoggerFactory.getLogger(HttpFunctions.class);

    //query aux-srv
    public static PickBackendResponse getBackendInstanceByRoomName(final String roomName) throws Exception {
        Instant methodStart = Instant.now();

        final int responseCode;
        final String responseJson;
        final PickBackendResponse responseObj;

        HttpURLConnection connection = null;

        try {
            final URL url = new URL(Constants.AUX_SRV_PICK_BACKEND_ENDPOINT + roomName);
            connection = (HttpURLConnection) url.openConnection();

            connection.setRequestProperty("Accept", "application/json");
            connection.setConnectTimeout(Constants.SOCKET_TIMEOUT_MS);
            connection.setReadTimeout(Constants.SOCKET_TIMEOUT_MS);

            responseCode = connection.getResponseCode();

            if (responseCode != 200) {
                Logging.logHttpError(String.format("!!! error while requesting aux-srv 'pick backend instance' - " +
                        "got non 200 response code: %s. RoomName: '%s'", responseCode, roomName));

                throw new RuntimeException();
            }

            try (BufferedReader br = new BufferedReader(
                    new InputStreamReader(connection.getInputStream(), StandardCharsets.UTF_8))) {
                StringBuilder response = new StringBuilder();

                String responseLine;
                while ((responseLine = br.readLine()) != null) {
                    response.append(responseLine.trim());
                }

                responseJson = response.toString();
            }

        } catch (final Exception e) {
            Logging.logHttpError(String.format("!!! error while requesting aux-srv 'pick backend instance'. RoomName: '%s'", roomName), e);

            throw e;
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }

        try {
            responseObj = Constants.OBJECT_MAPPER.readValue(responseJson, PickBackendResponse.class);
        } catch (final Exception e) {
            Logging.logHttpError(String.format("!!! error while parsing 'pick backend instance' response from aux-srv. RoomName: '%s'", roomName), e);

            throw e;
        }

        final long methodTiming = Duration.between(methodStart, Instant.now()).toMillis();

        if (StringUtils.isNotBlank(responseObj.getErrorMessage())) {
            Logging.logHttpError(String.format("!!! got error inside aux-srv 'pick backend instance' response: '%s', RoomName: '%s'. Request took %dms",
                    responseObj.getErrorMessage(), roomName, methodTiming));

            throw new RuntimeException();

        } else {
            Logging.logHttpInfo(String.format("response from aux-srv ('pick backend instance'): %dms, roomName: %s", methodTiming, roomName));

            return responseObj;
        }
    }

    public static Map<String, List<String>> getHomePage() throws Exception {
        Instant methodStart = Instant.now();

        final int responseCode;
        final Map<String, List<String>> responseHeaders;

        HttpURLConnection connection = null;

        try {
            final URL url = new URL(Constants.HTTP_PROTOCOL + Constants.SERVER_ROOT_ADDR);
            connection = (HttpURLConnection) url.openConnection();

            connection.setConnectTimeout(Constants.SOCKET_TIMEOUT_MS);
            connection.setReadTimeout(Constants.SOCKET_TIMEOUT_MS);

            responseCode = connection.getResponseCode();

            if (responseCode != 200) {
                Logging.logHttpError(String.format("!!! error while requesting home page - got non 200 response code: %s", responseCode));

                throw new RuntimeException();
            }

            responseHeaders = connection.getHeaderFields();

            //read response to fetch all data, but currently not used
            try (BufferedReader br = new BufferedReader(
                    new InputStreamReader(connection.getInputStream(), StandardCharsets.UTF_8))) {
                StringBuilder response = new StringBuilder();

                String responseLine;
                while ((responseLine = br.readLine()) != null) {
                    response.append(responseLine.trim());
                }
            }
        } catch (final Exception e) {
            Logging.logHttpError("!!! error while requesting aux-srv 'pick backend instance'", e);

            throw e;
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }

        final long methodTiming = Duration.between(methodStart, Instant.now()).toMillis();

        Logging.logHttpInfo(String.format("response from home page: %dms", methodTiming));

        return responseHeaders;
    }
}
