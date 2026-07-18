//! A from-scratch radix-2 Cooley-Tukey FFT (and inverse). No external crates.

use std::f64::consts::PI;

#[derive(Clone, Copy, Debug)]
pub struct Cpx {
    pub re: f64,
    pub im: f64,
}

impl Cpx {
    pub fn new(re: f64, im: f64) -> Self {
        Cpx { re, im }
    }
    pub fn add(self, o: Cpx) -> Cpx {
        Cpx::new(self.re + o.re, self.im + o.im)
    }
    pub fn sub(self, o: Cpx) -> Cpx {
        Cpx::new(self.re - o.re, self.im - o.im)
    }
    pub fn mul(self, o: Cpx) -> Cpx {
        Cpx::new(self.re * o.re - self.im * o.im, self.re * o.im + self.im * o.re)
    }
    pub fn mag(self) -> f64 {
        (self.re * self.re + self.im * self.im).sqrt()
    }
}

/// Smallest power of two >= n (min 1).
pub fn next_pow2(n: usize) -> usize {
    let mut p = 1;
    while p < n {
        p <<= 1;
    }
    p
}

/// Recursive radix-2 FFT. `input.len()` must be a power of two.
pub fn fft(input: &[Cpx]) -> Vec<Cpx> {
    let n = input.len();
    if n <= 1 {
        return input.to_vec();
    }
    let even: Vec<Cpx> = input.iter().step_by(2).copied().collect();
    let odd: Vec<Cpx> = input.iter().skip(1).step_by(2).copied().collect();
    let fe = fft(&even);
    let fo = fft(&odd);
    let mut out = vec![Cpx::new(0.0, 0.0); n];
    for k in 0..n / 2 {
        let ang = -2.0 * PI * (k as f64) / (n as f64);
        let tw = Cpx::new(ang.cos(), ang.sin()).mul(fo[k]);
        out[k] = fe[k].add(tw);
        out[k + n / 2] = fe[k].sub(tw);
    }
    out
}

/// Inverse FFT via the conjugation trick.
pub fn ifft(input: &[Cpx]) -> Vec<Cpx> {
    let n = input.len() as f64;
    let conj: Vec<Cpx> = input.iter().map(|c| Cpx::new(c.re, -c.im)).collect();
    let y = fft(&conj);
    y.iter().map(|c| Cpx::new(c.re / n, -c.im / n)).collect()
}

/// FFT of a real signal, zero-padded up to the next power of two.
pub fn fft_real(signal: &[f64]) -> Vec<Cpx> {
    let n = next_pow2(signal.len().max(1));
    let mut buf: Vec<Cpx> = signal.iter().map(|&x| Cpx::new(x, 0.0)).collect();
    buf.resize(n, Cpx::new(0.0, 0.0));
    fft(&buf)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn dominant_bin_of_cosine() {
        // cos at exactly k cycles over N samples -> energy peaks at bin k
        let n = 1024usize;
        let k = 16usize;
        let sig: Vec<f64> = (0..n)
            .map(|i| (2.0 * PI * (k as f64) * (i as f64) / (n as f64)).cos())
            .collect();
        let spec = fft_real(&sig);
        let mut best = 1;
        let mut best_mag = 0.0;
        for (i, c) in spec.iter().enumerate().take(n / 2).skip(1) {
            if c.mag() > best_mag {
                best_mag = c.mag();
                best = i;
            }
        }
        assert_eq!(best, k);
    }

    #[test]
    fn ifft_inverts_fft() {
        let sig: Vec<f64> = (0..8).map(|i| (i as f64) * 0.3 - 1.0).collect();
        let spec = fft_real(&sig);
        let back = ifft(&spec);
        for i in 0..8 {
            assert!((back[i].re - sig[i]).abs() < 1e-9, "i={i}");
        }
    }
}
