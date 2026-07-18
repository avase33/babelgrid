"use client";

import { useEffect, useRef, useState } from "react";

const SFU = process.env.NEXT_PUBLIC_SFU_URL || "ws://localhost:8080";
const ROOM = "call1";
const LANGS = ["es", "fr", "de"] as const;

type Transcript = {
  speaker: string;
  text: string;
  translations: Record<string, string>;
  dominant_hz: number;
  ts: number;
};

function float32ToInt16(buf: Float32Array): number[] {
  const out = new Array<number>(buf.length);
  for (let i = 0; i < buf.length; i++) {
    const s = Math.max(-1, Math.min(1, buf[i]));
    out[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
  }
  return out;
}

export default function Page() {
  const wsRef = useRef<WebSocket | null>(null);
  const ctxRef = useRef<AudioContext | null>(null);
  const procRef = useRef<ScriptProcessorNode | null>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const [status, setStatus] = useState("connecting");
  const [listening, setListening] = useState(false);
  const [target, setTarget] = useState<string>("es");
  const [speaker] = useState(() => "user_" + Math.random().toString(36).slice(2, 5));
  const [feed, setFeed] = useState<Transcript[]>([]);

  useEffect(() => {
    const ws = new WebSocket(
      `${SFU}/ws?room=${ROOM}&speaker=${speaker}&lang=en`,
    );
    wsRef.current = ws;
    ws.onopen = () => setStatus("live");
    ws.onclose = () => setStatus("offline");
    ws.onerror = () => setStatus("offline");
    ws.onmessage = (e: MessageEvent) => {
      try {
        const msg = JSON.parse(e.data);
        if (msg.type === "transcript") {
          setFeed((f) => [msg as Transcript, ...f].slice(0, 40));
        }
      } catch {
        /* ignore */
      }
    };
    return () => ws.close();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const start = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;
      const ctx = new AudioContext();
      ctxRef.current = ctx;
      const source = ctx.createMediaStreamSource(stream);
      const proc = ctx.createScriptProcessor(4096, 1, 1);
      procRef.current = proc;
      proc.onaudioprocess = (ev: AudioProcessingEvent) => {
        const ws = wsRef.current;
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        const pcm = float32ToInt16(ev.inputBuffer.getChannelData(0));
        ws.send(
          JSON.stringify({
            type: "audio",
            room: ROOM,
            speaker,
            lang: "en",
            sr: Math.round(ctx.sampleRate),
            pcm,
          }),
        );
      };
      source.connect(proc);
      proc.connect(ctx.destination);
      setListening(true);
    } catch {
      setStatus("mic-denied");
    }
  };

  const stop = () => {
    procRef.current?.disconnect();
    ctxRef.current?.close();
    streamRef.current?.getTracks().forEach((t) => t.stop());
    procRef.current = null;
    ctxRef.current = null;
    streamRef.current = null;
    setListening(false);
  };

  return (
    <main style={{ padding: 24, maxWidth: 900, margin: "0 auto" }}>
      <header style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
        <h1 style={{ fontSize: 22, margin: 0 }}>babelgrid</h1>
        <span style={{ fontSize: 13, color: status === "live" ? "#4ec9b0" : "#ff7b72" }}>
          {status} · {speaker}
        </span>
      </header>
      <p style={{ color: "#8b949e", fontSize: 13 }}>
        speak into the mic — audio is cleaned by the Rust DSP, transcribed, and
        translated for every listener. (Offline STT is a deterministic mock.)
      </p>

      <div style={{ display: "flex", gap: 10, alignItems: "center", marginTop: 8 }}>
        {!listening ? (
          <button onClick={start} style={btn("#238636")}>● start mic</button>
        ) : (
          <button onClick={stop} style={btn("#6e2c2c")}>■ stop</button>
        )}
        <label style={{ fontSize: 13, color: "#8b949e" }}>
          show translation:&nbsp;
          <select
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            style={{ background: "#0d1117", color: "#e6edf3", border: "1px solid #30363d", borderRadius: 6, padding: "4px 6px" }}
          >
            {LANGS.map((l) => (
              <option key={l} value={l}>{l}</option>
            ))}
          </select>
        </label>
      </div>

      <section style={{ marginTop: 20, border: "1px solid #30363d", borderRadius: 10 }}>
        {feed.length === 0 && (
          <p style={{ padding: 16, color: "#8b949e" }}>no transcripts yet…</p>
        )}
        {feed.map((t, i) => (
          <div
            key={`${t.ts}-${i}`}
            style={{ padding: "10px 14px", borderTop: i ? "1px solid #21262d" : "none" }}
          >
            <div style={{ display: "flex", justifyContent: "space-between" }}>
              <b style={{ color: "#58a6ff" }}>{t.speaker}</b>
              <span style={{ color: "#8b949e", fontSize: 12 }}>{Math.round(t.dominant_hz)} Hz</span>
            </div>
            <div style={{ marginTop: 2 }}>{t.text}</div>
            <div style={{ color: "#4ec9b0", fontSize: 14, marginTop: 2 }}>
              [{target}] {t.translations?.[target] ?? "…"}
            </div>
          </div>
        ))}
      </section>
    </main>
  );
}

function btn(bg: string): React.CSSProperties {
  return {
    padding: "8px 14px",
    background: bg,
    color: "#fff",
    border: "none",
    borderRadius: 8,
    cursor: "pointer",
    fontSize: 14,
  };
}
