# Wiki Index

Technical and engineering knowledge base. Updated on every ingest.

## How to Use
- **Adding a source:** Drop a URL, PDF, or image into the chat with a brief note. Andy will ingest it fully.
- **Querying:** Ask any question — Andy reads this index first, then drills into relevant pages.
- **Linting:** Ask "run a wiki lint" to health-check for contradictions, orphans, and gaps.

---

## Pages

<!-- Andy maintains this section. Format: - [Page Title](path) — one-line summary -->

### 🔬 Research
- [Best Ollama Model for Java Code Review (Apr 2026)](ollama-java-code-review/summary.md) — **Current picks**: Qwen3.5-27B (overall #1), Devstral Small 2 (agentic), Qwen2.5-Coder-14B Q8_0 (Java benchmarks)
- [Qwen3.5 vs Qwen-Coder: Java Code Review](ollama-java-code-review/qwen35-vs-coder.md) — **Updated pick**: Qwen3.5-27B INT4 beats Qwen2.5-Coder-32B for review tasks; no Qwen3.5-Coder exists
- [Gemma 4 vs Qwen3.5-27B vs Qwen2.5-Coder-14B: Java Code Review](ollama-java-code-review/gemma4-comparison.md) — Gemma 4 NOT recommended for Java review (no SWE-bench/Java benchmarks published); Qwen3.5-27B still top pick
- [Kimi K2.5 vs Wiki Models: Java Code Review](ollama-java-code-review/kimi-k2.5-comparison.md) — Kimi K2.5 DISQUALIFIED for local 20GB GPU (1T params, ~580 GB GGUF; Ollama tag is cloud-only)
- [Model List Comparison: 7 Candidate Models](ollama-java-code-review/model-list-comparison.md) — Qwen2.5-Coder-32B best of the 7 but ~2GB over limit; DeepSeek-Coder-V2-Lite/CodeLlama/Mistral not recommended
- [Vibe Coding on Steam Deck with Claude CLI + Ollama](ollama-java-code-review/steam-deck-vibe-coding.md) — Top pick: qwen2.5-coder:7b (4.7GB, ~84% HumanEval); ROCm workaround via HSA_OVERRIDE_GFX_VERSION=10.3.0
- [Article: dev.to Ollama Cloud Code Review Comparison](ollama-java-code-review/article-dev-to-comparison.md) — DeepSeek v3.1 wins (4.25/5) over Qwen 3.5 (3.8/5) and GPT-OSS (2.9/5) on Python PRs; cloud models only
- [GLM-Z1 / GLM-5.1 vs Wiki Models: Java Code Review](ollama-java-code-review/glm-z1-comparison.md) — GLM-5.1 cloud-only (744B); local GLM-Z1-32B only ~33% SWE-bench, no Java benchmarks; NOT RECOMMENDED over Devstral Small 2

### 🏷️ Entities
- [Qwen2.5-Coder](entities/qwen2.5-coder.md) — Alibaba coding LLM family; 14B/32B variants; top open-source code benchmarks
- [Qwen3-Coder](entities/qwen3-coder.md) — 2026 MoE coding model; 30B/3.3B active; SWE-bench 69.6%
- [Qwen3.5](entities/qwen3.5.md) — Feb 2026 multimodal general model; 262K context; SWE-bench 72.4%; no Coder variant
- [Ollama](entities/ollama.md) — Local LLM runtime with GPU offloading and REST API
- [Gemma 4](entities/gemma-4.md) — Google Apr 2026 open-weight multimodal family; 26B MoE/31B Dense; Apache 2.0; no SWE-bench published; no code variant
- [Kimi K2.5](entities/kimi-k2.5.md) — Moonshot AI Feb 2026; 1T-param MoE/32B active; 256K ctx; SWE-bench 76.8%; cloud-only on Ollama; Modified MIT license
- [Qwen3-Coder-Next](entities/qwen3-coder-next.md) — Alibaba Jan 2026; 80B total/3B active MoE; 256K ctx; SWE-bench 70.6%; ~52GB VRAM — does NOT fit 20GB GPU; Apache 2.0
- [Devstral Small 2](entities/devstral-small-2.md) — Mistral AI Nov 2025; 24B dense; 256K ctx; SWE-bench 65.8%; fits 20GB GPU at Q4_K_M (~14GB); agentic coding specialist; Apache 2.0
- [Devstral 2](entities/devstral-2.md) — Mistral AI Nov 2025; 123B dense; 256K ctx; SWE-bench 72.2%; ~75GB Q4 — does NOT fit 20GB GPU; Mistral Research License
- [GLM-Z1 / GLM-4.7 / GLM-5.x](entities/glm-z1.md) — Z.AI (Zhipu/Tsinghua); MIT license; local: Z1-32B (~20GB) or 4.7-Flash (16GB, 200K ctx); cloud: GLM-4.7 73.8% / GLM-5 77.8% SWE-bench; local models lack published coding benchmarks

### 💡 Concepts
- [LLM Quantization](concepts/llm-quantization.md) — GGUF quantization levels, VRAM estimates, Q4/Q5/Q8 tradeoffs
- [Spring Boot Maven Plugin — Docker Image Builds](concepts/spring-boot-maven-docker.md) — Build OCI images with `mvn spring-boot:build-image`; Cloud Native Buildpacks / Paketo; custom name, platform, registry push, Java version config

### 🏗️ Infrastructure
- [Wiki UI (MkDocs)](infrastructure/wiki-ui.md) — How the read-only web UI is built and served; systemd timer + static file server architecture

---

## Stats

- Sources ingested: 7 (web research sessions + 1 article)
- Wiki pages: 22
- Last updated: 2026-04-15

