package main.util;

import main.domain.OutMessageFrame;

import java.util.*;

public class Util {

    public static void sleep(long delay) {
        try {
            Thread.sleep(delay);
        } catch (final InterruptedException e) {
            //ignore
        }
    }

    public static Optional<OutMessageFrame> findMessageByExpectedCommandAndRequestId(
            final String command,
            final String requestId,
            final List<OutMessageFrame> messages
    ) {
        return messages.stream()
                .filter(message -> command.equals(message.getCommand()) && requestId.equals(message.getRequestId()))
                .findAny();
    }

    public static <K, V extends Comparable<? super V>> Map<K, V> sortByValue(Map<K, V> map) {
        List<Map.Entry<K, V>> list = new ArrayList<>(map.entrySet());
        list.sort(Map.Entry.comparingByValue());

        Map<K, V> result = new LinkedHashMap<>();
        for (Map.Entry<K, V> entry : list) {
            result.put(entry.getKey(), entry.getValue());
        }

        return result;
    }
}
