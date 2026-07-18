//! HTTP front-end for the DSP core (axum).
//!   POST /process  { pcm:[i16], sr:u32 }  -> Features
//!   GET  /healthz

use axum::{routing::{get, post}, Json, Router};
use babelgrid_dsp::{process, Features};
use serde::Deserialize;

#[derive(Deserialize)]
struct ProcessRequest {
    pcm: Vec<i16>,
    #[serde(default = "default_sr")]
    sr: u32,
}

fn default_sr() -> u32 {
    16000
}

async fn process_handler(Json(req): Json<ProcessRequest>) -> Json<Features> {
    Json(process(&req.pcm, req.sr))
}

async fn healthz() -> Json<serde_json::Value> {
    Json(serde_json::json!({ "status": "ok" }))
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/process", post(process_handler))
        .route("/healthz", get(healthz));

    let addr = std::env::var("BABELGRID_DSP_ADDR").unwrap_or_else(|_| "0.0.0.0:8092".into());
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    println!("babelgrid-dsp listening on {addr}");
    axum::serve(listener, app).await.unwrap();
}
