package main;


import main.util.Util;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.concurrent.atomic.AtomicInteger;

public class Main {
    private static final Logger LOGGER = LoggerFactory.getLogger(Main.class);

    public static final int SPAWNER_THREADS_COUNT = 50;   //50

    public static final int TEST_ROOMS_AMOUNT = 10;
    public static final int TEST_ROOMS_RECREATE_COUNT = 1;

    public static final int ROOM_LIFE_SPAN_MS = 15 * 60 * 1000;        //4 * 60 * 1000

    public static final int ROOM_USERS_ADDING_STEPS_NUM = 2;         //5
    public static final int ROOM_USERS_ADDING_AMOUNT_PER_STEP = 2;  //10

    public static final int ROOM_USERS_ADDING_STEP_DELAY_MS = 1000; //10000

    public static boolean stopped = false;

    public static void main(String[] args) {
        final AtomicInteger nextRoomNumber = new AtomicInteger(0);
        final List<Thread> roomThreads = new ArrayList<>();

        //make X 'spawner' threads
        for (int i = 0; i < SPAWNER_THREADS_COUNT; i++) {
            LOGGER.info(String.format("=== creating new 'spawner' thread #'%s'\n\n", i));

            final Thread spawnerThread = new Thread(() -> {
                //spawner thread spawns new 'room' threads
                for (int ii = 0; ii < TEST_ROOMS_AMOUNT; ii++) {
                    Util.sleep(
                            new Random().nextInt(1000) + 1000
                    );

                    final Thread roomThread = new Thread(() -> {
                        int roomNumber = nextRoomNumber.incrementAndGet();
                        int roomRecreatedCount = 0;

                        try {
                            while (!stopped && roomRecreatedCount++ < TEST_ROOMS_RECREATE_COUNT) {
                                new TestRoom().startTestRoom(roomNumber);

                                roomNumber = nextRoomNumber.incrementAndGet();
                            }
                        } catch (Exception e) {
                            LOGGER.error(String.format("error from room %d", roomNumber), e);
                        }
                    });

                    roomThread.start();

                    roomThreads.add(roomThread);
                }

            });

            spawnerThread.start();

            Util.sleep(
                    2000 + new Random().nextInt(3000)
            );
        }

        Util.sleep(
                Integer.MAX_VALUE
        );

//        try {
//            for (final Thread thread : roomThreads) {
//                thread.join();
//            }
//        } catch (InterruptedException e) {
//            throw new RuntimeException(e);
//        }
    }
}
