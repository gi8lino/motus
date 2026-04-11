import assert from "node:assert/strict";
import test from "node:test";

import { createAudioController } from "../src/utils/audioController.ts";

type PlayGate = {
  promise: Promise<void>;
  resolve: () => void;
  reject: (error?: unknown) => void;
};

function createGate(): PlayGate {
  let resolve!: () => void;
  let reject!: (error?: unknown) => void;
  const promise = new Promise<void>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

class FakeAudio {
  static instances: FakeAudio[] = [];
  static playGates: PlayGate[] = [];

  src: string;
  preload = "";
  playsInline = false;
  muted = false;
  currentTime = 0;
  pauseCalls = 0;
  loadCalls = 0;
  playCalls = 0;
  playSnapshots: Array<{
    muted: boolean;
    playsInline: boolean;
    preload: string;
    src: string;
  }> = [];

  constructor(src = "") {
    this.src = src;
    FakeAudio.instances.push(this);
  }

  static reset() {
    FakeAudio.instances = [];
    FakeAudio.playGates = [];
  }

  addEventListener(): void {
    // No-op for controller tests.
  }

  play(): Promise<void> {
    this.playCalls += 1;
    this.playSnapshots.push({
      muted: this.muted,
      playsInline: this.playsInline,
      preload: this.preload,
      src: this.src,
    });

    const gate = FakeAudio.playGates.shift();
    return gate ? gate.promise : Promise.resolve();
  }

  pause(): void {
    this.pauseCalls += 1;
  }

  load(): void {
    this.loadCalls += 1;
  }
}

function installFakeAudio() {
  FakeAudio.reset();

  const globalWithAudio = globalThis as typeof globalThis & {
    Audio?: typeof Audio;
  };
  const originalAudio = globalWithAudio.Audio;

  globalWithAudio.Audio = FakeAudio as unknown as typeof Audio;

  return () => {
    if (originalAudio) globalWithAudio.Audio = originalAudio;
    else delete globalWithAudio.Audio;
    FakeAudio.reset();
  };
}

test("unlock primes a muted inline audio element for mobile playback", async () => {
  const restore = installFakeAudio();

  try {
    const controller = createAudioController();
    await controller.unlock();

    assert.equal(FakeAudio.instances.length, 1);

    const [audio] = FakeAudio.instances;
    assert.equal(audio.playCalls, 1);
    assert.equal(audio.playSnapshots[0]?.muted, true);
    assert.equal(audio.playSnapshots[0]?.playsInline, true);
    assert.equal(audio.playSnapshots[0]?.preload, "auto");
    assert.match(audio.playSnapshots[0]?.src || "", /^data:audio\/wav;base64,/);
    assert.equal(audio.pauseCalls, 1);
    assert.equal(audio.loadCalls, 1);
    assert.equal(audio.src, "");
  } finally {
    restore();
  }
});

test("play waits for an in-flight unlock before starting sound playback", async () => {
  const restore = installFakeAudio();

  try {
    const unlockGate = createGate();
    FakeAudio.playGates.push(unlockGate);

    const controller = createAudioController();
    const unlockPromise = controller.unlock();

    controller.play("/sounds/beep.wav");

    assert.equal(FakeAudio.instances.length, 2);

    const soundAudio = FakeAudio.instances[1];
    assert.equal(soundAudio.playCalls, 0);

    unlockGate.resolve();
    await unlockPromise;
    await Promise.resolve();

    assert.equal(soundAudio.playCalls, 1);
    assert.equal(soundAudio.playSnapshots[0]?.playsInline, true);
    assert.equal(soundAudio.playSnapshots[0]?.preload, "auto");
    assert.equal(soundAudio.playSnapshots[0]?.src, "/sounds/beep.wav");
  } finally {
    restore();
  }
});

test("pause cancels playback queued behind an unfinished unlock", async () => {
  const restore = installFakeAudio();

  try {
    const unlockGate = createGate();
    FakeAudio.playGates.push(unlockGate);

    const controller = createAudioController();
    void controller.unlock();
    controller.play("/sounds/beep.wav");
    controller.pause();

    const soundAudio = FakeAudio.instances[1];

    unlockGate.resolve();
    await Promise.resolve();
    await Promise.resolve();

    assert.equal(soundAudio.playCalls, 0);
  } finally {
    restore();
  }
});
