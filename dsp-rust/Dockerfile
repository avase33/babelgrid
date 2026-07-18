FROM rust:1.79-slim AS build
WORKDIR /app
COPY Cargo.toml ./
COPY src ./src
RUN cargo build --release --bin babelgrid-dsp

FROM debian:bookworm-slim
COPY --from=build /app/target/release/babelgrid-dsp /usr/local/bin/babelgrid-dsp
ENV BABELGRID_DSP_ADDR=0.0.0.0:8092
EXPOSE 8092
CMD ["babelgrid-dsp"]
