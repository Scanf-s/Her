// AudioWorklet processor: runs on the audio rendering thread (separate from the UI).
// For every render block it forwards the mono PCM frame to the main thread, where
// it is accumulated and later encoded to a WAV file.
//
// We do NOT write to `outputs`, so the engine emits silence downstream (no echo),
// while connecting this node toward destination keeps it pulled/processed.
class PCMRecorder extends AudioWorkletProcessor {
  process(inputs) {
    const channel = inputs[0][0]; // first input, first (mono) channel: Float32Array
    if (channel) {
      // The engine reuses the buffer, so copy before posting.
      this.port.postMessage(channel.slice(0));
    }
    return true; // keep the processor alive
  }
}

registerProcessor('pcm-recorder', PCMRecorder);
