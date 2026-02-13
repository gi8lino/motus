import assert from "node:assert/strict";
import test from "node:test";

import { createTrainingKeyboardHandler } from "../../src/hooks/trainingKeyboard/handler.ts";

type EventInput = {
  code: string;
  repeat?: boolean;
  tagName?: string;
};

function makeEvent(input: EventInput) {
  let prevented = false;
  const event = {
    code: input.code,
    repeat: Boolean(input.repeat),
    preventDefault() {
      prevented = true;
    },
    target: input.tagName ? { tagName: input.tagName } : null,
  };
  return { event: event as unknown as KeyboardEvent, prevented: () => prevented };
}

function createArgs(overrides?: Partial<Parameters<typeof createTrainingKeyboardHandler>[0]>) {
  const calls = {
    overrunPostpone: 0,
    overrunPause: 0,
    pause: 0,
    start: 0,
    stopAudio: 0,
    next: 0,
    finish: 0,
    setSummary: [] as Array<string | null>,
  };

  const args: Parameters<typeof createTrainingKeyboardHandler>[0] = {
    trainingRef: {
      current: {
        trainingId: "t1",
        workoutId: "w1",
        userId: "u1",
        currentIndex: 0,
        running: true,
        done: false,
        startedAt: new Date().toISOString(),
        steps: [{ id: "s1", type: "set", name: "Step 1" }],
      },
    },
    overrunModalRef: { current: null },
    handleOverrunPostpone: () => {
      calls.overrunPostpone += 1;
    },
    handleOverrunPause: () => {
      calls.overrunPause += 1;
    },
    handlePause: () => {
      calls.pause += 1;
    },
    handleStart: () => {
      calls.start += 1;
    },
    stopActiveAudio: () => {
      calls.stopAudio += 1;
    },
    onNext: () => {
      calls.next += 1;
    },
    onFinishTraining: async () => {
      calls.finish += 1;
      return "summary";
    },
    setFinishSummary: (summary) => {
      calls.setSummary.push(summary);
    },
    ...overrides,
  };

  return { args, calls };
}

test("overrun modal handles enter and space shortcuts", () => {
  const { args, calls } = createArgs({
    overrunModalRef: { current: { show: true } },
  });
  const handler = createTrainingKeyboardHandler(args);

  const enter = makeEvent({ code: "Enter" });
  handler(enter.event);
  assert.equal(enter.prevented(), true);
  assert.equal(calls.overrunPostpone, 1);

  const space = makeEvent({ code: "Space" });
  handler(space.event);
  assert.equal(space.prevented(), true);
  assert.equal(calls.overrunPause, 1);
});

test("space toggles pause when training is running", () => {
  const { args, calls } = createArgs();
  const handler = createTrainingKeyboardHandler(args);

  const event = makeEvent({ code: "Space" });
  handler(event.event);

  assert.equal(event.prevented(), true);
  assert.equal(calls.pause, 1);
  assert.equal(calls.start, 0);
});

test("space toggles start when training is paused", () => {
  const { args, calls } = createArgs({
    trainingRef: {
      current: {
        trainingId: "t1",
        workoutId: "w1",
        userId: "u1",
        currentIndex: 0,
        running: false,
        done: false,
        startedAt: new Date().toISOString(),
        steps: [{ id: "s1", type: "set", name: "Step 1" }],
      },
    },
  });
  const handler = createTrainingKeyboardHandler(args);

  const event = makeEvent({ code: "Space" });
  handler(event.event);

  assert.equal(event.prevented(), true);
  assert.equal(calls.pause, 0);
  assert.equal(calls.start, 1);
});

test("enter advances to next step when not on last step", () => {
  const { args, calls } = createArgs({
    trainingRef: {
      current: {
        trainingId: "t1",
        workoutId: "w1",
        userId: "u1",
        currentIndex: 0,
        running: true,
        done: false,
        startedAt: new Date().toISOString(),
        steps: [
          { id: "s1", type: "set", name: "Step 1" },
          { id: "s2", type: "set", name: "Step 2" },
        ],
      },
    },
  });
  const handler = createTrainingKeyboardHandler(args);

  const event = makeEvent({ code: "Enter" });
  handler(event.event);

  assert.equal(event.prevented(), true);
  assert.equal(calls.stopAudio, 1);
  assert.equal(calls.next, 1);
  assert.equal(calls.finish, 0);
});

test("enter finishes on last step and forwards summary", async () => {
  const { args, calls } = createArgs({
    trainingRef: {
      current: {
        trainingId: "t1",
        workoutId: "w1",
        userId: "u1",
        currentIndex: 1,
        running: true,
        done: false,
        startedAt: new Date().toISOString(),
        steps: [
          { id: "s1", type: "set", name: "Step 1" },
          { id: "s2", type: "set", name: "Step 2" },
        ],
      },
    },
  });
  const handler = createTrainingKeyboardHandler(args);

  const event = makeEvent({ code: "Enter" });
  handler(event.event);

  await Promise.resolve();
  await Promise.resolve();

  assert.equal(event.prevented(), true);
  assert.equal(calls.stopAudio, 1);
  assert.equal(calls.finish, 1);
  assert.deepEqual(calls.setSummary, ["summary"]);
});

test("keyboard shortcuts are ignored for repeated keys and input fields", () => {
  const { args, calls } = createArgs();
  const handler = createTrainingKeyboardHandler(args);

  handler(makeEvent({ code: "Space", repeat: true }).event);
  handler(makeEvent({ code: "Space", tagName: "INPUT" }).event);

  assert.equal(calls.pause, 0);
  assert.equal(calls.start, 0);
  assert.equal(calls.next, 0);
  assert.equal(calls.finish, 0);
});
