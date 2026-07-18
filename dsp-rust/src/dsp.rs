//! The DSP layer: takes raw PCM, runs a spectral noise gate, and extracts the
//! compact features the (mock) STT layer needs — energy, whether the frame is
//! voiced, its dominant frequency, and one stable token per voiced segment.

use crate::fft::{fft_real, ifft, Cpx};
use serde::Serialize;

#[derive(Serialize, Debug)]
pub struct Features {
    pub rms: f64,
    pub rms_gated: f64,
    pub voiced: bool,
    pub dominant_hz: f64,
    pub duration_ms: f64,
    pub tokens: Vec<u32>,
}

const VOICE_RMS_THRESHOLD: f64 = 0.02;

fn rms(x: &[f64]) -> f64 {
    if x.is_empty() {
        return 0.0;
    }
    (x.iter().map(|v| v * v).sum::<f64>() / x.len() as f64).sqrt()
}

/// Spectral noise gate: zero every frequency bin below a fraction of the peak
/// magnitude, then invert back to the time domain.
fn spectral_gate(signal: &[f64], frac: f64) -> Vec<f64> {
    if signal.is_empty() {
        return vec![];
    }
    let mut spec = fft_real(signal);
    let peak = spec.iter().map(|c| c.mag()).fold(0.0_f64, f64::max);
    let thr = peak * frac;
    for c in spec.iter_mut() {
        if c.mag() < thr {
            *c = Cpx::new(0.0, 0.0);
        }
    }
    let back = ifft(&spec);
    back.iter().take(signal.len()).map(|c| c.re).collect()
}

/// FNV-1a style token from a voiced segment's shape.
fn token(seg_index: usize, len_frames: usize, energy_milli: u64) -> u32 {
    let mut h: u32 = 2166136261;
    for v in [seg_index as u32, len_frames as u32, energy_milli as u32] {
        h ^= v;
        h = h.wrapping_mul(16777619);
    }
    h % 4096
}

/// Analyse one PCM frame.
pub fn process(pcm: &[i16], sr: u32) -> Features {
    let x: Vec<f64> = pcm.iter().map(|&s| s as f64 / 32768.0).collect();
    let overall_rms = rms(&x);

    let gated = spectral_gate(&x, 0.10);
    let rms_gated = rms(&gated);

    // dominant frequency from the (pre-gate) spectrum
    let spec = fft_real(&x);
    let n = spec.len().max(1);
    let mut best_bin = 0usize;
    let mut best_mag = 0.0;
    for (i, c) in spec.iter().enumerate().take(n / 2).skip(1) {
        if c.mag() > best_mag {
            best_mag = c.mag();
            best_bin = i;
        }
    }
    let dominant_hz = best_bin as f64 * sr as f64 / n as f64;

    // voiced-segment tokenisation over ~10 ms frames
    let frame = (sr as usize / 100).max(1);
    let mut tokens = Vec::new();
    let mut seg_len = 0usize;
    let mut seg_energy = 0.0f64;
    let mut seg_index = 0usize;
    let rel = (overall_rms * 0.3).max(0.01);
    let flush = |tokens: &mut Vec<u32>, idx: usize, len: usize, energy: f64| {
        if len > 0 {
            let avg = energy / len as f64;
            tokens.push(token(idx, len, (avg * 1000.0) as u64));
        }
    };
    for chunk in x.chunks(frame) {
        let e = rms(chunk);
        if e > rel {
            seg_len += 1;
            seg_energy += e;
        } else if seg_len > 0 {
            flush(&mut tokens, seg_index, seg_len, seg_energy);
            seg_index += 1;
            seg_len = 0;
            seg_energy = 0.0;
        }
    }
    flush(&mut tokens, seg_index, seg_len, seg_energy);

    let voiced = overall_rms > VOICE_RMS_THRESHOLD;
    if voiced && tokens.is_empty() {
        tokens.push(token(0, 1, (overall_rms * 1000.0) as u64));
    }

    Features {
        rms: (overall_rms * 1000.0).round() / 1000.0,
        rms_gated: (rms_gated * 1000.0).round() / 1000.0,
        voiced,
        dominant_hz: (dominant_hz * 10.0).round() / 10.0,
        duration_ms: (pcm.len() as f64 / sr as f64 * 1000.0 * 10.0).round() / 10.0,
        tokens,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::f64::consts::PI;

    fn sine(freq: f64, sr: u32, n: usize, amp: f64) -> Vec<i16> {
        (0..n)
            .map(|i| {
                let v = amp * (2.0 * PI * freq * i as f64 / sr as f64).sin();
                (v * 32767.0) as i16
            })
            .collect()
    }

    #[test]
    fn detects_dominant_frequency() {
        let pcm = sine(250.0, 16000, 1024, 0.8);
        let f = process(&pcm, 16000);
        assert!((f.dominant_hz - 250.0).abs() < 20.0, "got {}", f.dominant_hz);
        assert!(f.voiced);
        assert!(!f.tokens.is_empty());
    }

    #[test]
    fn silence_is_unvoiced() {
        let pcm = vec![0i16; 512];
        let f = process(&pcm, 16000);
        assert!(!f.voiced);
        assert!(f.tokens.is_empty());
        assert_eq!(f.rms, 0.0);
    }

    #[test]
    fn tokens_are_deterministic() {
        let pcm = sine(180.0, 16000, 1600, 0.6);
        let a = process(&pcm, 16000);
        let b = process(&pcm, 16000);
        assert_eq!(a.tokens, b.tokens);
    }

    #[test]
    fn gate_does_not_amplify() {
        let pcm = sine(300.0, 16000, 1024, 0.5);
        let f = process(&pcm, 16000);
        // gating removes energy, never adds it
        assert!(f.rms_gated <= f.rms + 1e-6);
    }
}
