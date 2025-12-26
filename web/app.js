// secondbrain frontend (no build)
// Data flow:
// - CreateNote: POST JSON to /api/CreateNote
// - QueryNotes: POST JSON to /api/QueryNotes and parse SSE-like stream:
//     event: response  -> append chunk to assistant message
//     event: done      -> finalize
//     event: error     -> show error

const el = (id) => document.getElementById(id);

const statusBadge = el("statusBadge");

const noteForm = el("noteForm");
const noteId = el("noteId");
const noteContent = el("noteContent");
const noteResult = el("noteResult");
const noteClearBtn = el("noteClearBtn");

const chatLog = el("chatLog");
const chatForm = el("chatForm");
const chatInput = el("chatInput");
const numResultsInput = el("numResults");
const sendBtn = el("sendBtn");
const stopBtn = el("stopBtn");
const chatHint = el("chatHint");

let activeAbortController = null;

// -----------------------------
// UI helpers
// -----------------------------
function setStatus(text) {
    statusBadge.textContent = text;
}

function scrollChatToBottom() {
    chatLog.scrollTop = chatLog.scrollHeight;
}

function addMessage({ role, text, variant = role }) {
    const row = document.createElement("div");
    row.className = "msg";

    const roleEl = document.createElement("div");
    roleEl.className = "role";
    roleEl.textContent = role;

    const bubble = document.createElement("div");
    bubble.className = `bubble ${variant}`;
    bubble.textContent = text;

    row.appendChild(roleEl);
    row.appendChild(bubble);
    chatLog.appendChild(row);
    scrollChatToBottom();

    return bubble; // return bubble so we can update it during streaming
}

function setNoteResult(msg, isError = false) {
    noteResult.textContent = msg;
    noteResult.style.color = isError ? "rgba(248,113,113,0.95)" : "rgba(110,231,183,0.95)";
    if (!msg) noteResult.style.color = "";
}

function setChatHint(msg) {
    chatHint.textContent = msg;
}

// -----------------------------
// API helpers
// -----------------------------
async function postJSON(url, body, { signal } = {}) {
    const res = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
        signal,
    });

    // CreateNote might respond JSON or text
    const contentType = res.headers.get("content-type") || "";
    const text = await res.text();

    if (!res.ok) {
        throw new Error(text || `Request failed (${res.status})`);
    }

    if (contentType.includes("application/json")) {
        try {
            return JSON.parse(text);
        } catch {
            // fallthrough to raw text
        }
    }
    return text;
}

// Parse an SSE-like stream from fetch Response.body.
// We handle frames separated by blank lines; each frame has lines like:
//   event: response
//   data: hello
// data may appear multiple times; we join with '\n' (SSE semantics).
async function consumeSSELikeStream(response, handlers, { signal } = {}) {
    if (!response.ok) {
        const errText = await response.text().catch(() => "");
        throw new Error(errText || `Request failed (${response.status})`);
    }
    if (!response.body) {
        throw new Error("Streaming not supported by this browser/response.");
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder("utf-8");

    let buffer = "";

    const emitFrame = (frameText) => {
        // parse one frame
        const lines = frameText.split(/\r?\n/);
        let eventName = "message";
        const dataLines = [];

        for (const line of lines) {
            if (!line || line.startsWith(":")) continue;
            const idx = line.indexOf(":");
            if (idx === -1) continue;

            const field = line.slice(0, idx).trim();
            const value = line.slice(idx + 1).trimStart();

            if (field === "event") eventName = value;
            if (field === "data") dataLines.push(value);
        }

        const data = dataLines.join("\n");

        if (eventName === "response" && handlers.onResponse) handlers.onResponse(data);
        else if (eventName === "done" && handlers.onDone) handlers.onDone(data);
        else if (eventName === "error" && handlers.onError) handlers.onError(data);
        else if (handlers.onMessage) handlers.onMessage({ event: eventName, data });
    };

    while (true) {
        if (signal?.aborted) {
            // Allow caller to treat abort as a normal exit
            return;
        }

        const { value, done } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });

        // frames separated by blank line
        let splitIndex;
        while ((splitIndex = buffer.search(/\r?\n\r?\n/)) !== -1) {
            const frame = buffer.slice(0, splitIndex);
            buffer = buffer.slice(splitIndex).replace(/^\r?\n\r?\n/, "");
            if (frame.trim()) emitFrame(frame);
        }
    }

    // flush any remaining text as a last frame (optional)
    if (buffer.trim()) emitFrame(buffer);
}

// -----------------------------
// CreateNote wiring
// -----------------------------
noteForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    setNoteResult("");

    const content = noteContent.value.trim();
    const id = noteId.value.trim();

    if (!content) {
        setNoteResult("Note content is required.", true);
        return;
    }

    setStatus("Saving note…");
    try {
        const payload = { content };
        if (id) payload.id = id;

        const result = await postJSON("/api/CreateNote", payload);
        // result could be json or text
        const msg = typeof result === "string" ? result : `Saved note${result?.id ? ` (${result.id})` : ""}.`;
        setNoteResult(msg || "Saved.");
        noteContent.value = "";
        noteId.value = "";
    } catch (err) {
        setNoteResult(err?.message || "Failed to save note.", true);
    } finally {
        setStatus("Idle");
    }
});

noteClearBtn.addEventListener("click", () => {
    noteId.value = "";
    noteContent.value = "";
    setNoteResult("");
});

// -----------------------------
// Chat / QueryNotes wiring
// -----------------------------
function setStreamingUI(isStreaming) {
    sendBtn.disabled = isStreaming;
    stopBtn.disabled = !isStreaming;
    chatInput.disabled = isStreaming;
    numResultsInput.disabled = isStreaming;
}

stopBtn.addEventListener("click", () => {
    if (activeAbortController) {
        activeAbortController.abort();
    }
});

chatForm.addEventListener("submit", async (e) => {
    e.preventDefault();

    const question = chatInput.value.trim();
    if (!question) return;

    // Abort any previous stream (defensive)
    if (activeAbortController) {
        activeAbortController.abort();
        activeAbortController = null;
    }

    const numResultsRaw = Number(numResultsInput.value);
    const numResults = Number.isFinite(numResultsRaw) ? Math.max(1, Math.min(50, numResultsRaw)) : 5;

    addMessage({ role: "user", text: question, variant: "user" });
    chatInput.value = "";

    // Create placeholder assistant bubble we will fill incrementally
    const assistantBubble = addMessage({ role: "assistant", text: "", variant: "assistant" });

    const ac = new AbortController();
    activeAbortController = ac;

    setStreamingUI(true);
    setStatus("Thinking…");

    try {
        const res = await fetch("/api/QueryNotes", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ text: question, numResults }),
            signal: ac.signal,
        });

        let built = "";

        await consumeSSELikeStream(
            res,
            {
                onResponse: (chunk) => {
                    // Append partial tokens/chunks
                    built += chunk + " ";
                    assistantBubble.textContent = built;
                    scrollChatToBottom();
                },
                onDone: () => {
                    setChatHint("Done.");
                },
                onError: (msg) => {
                    assistantBubble.classList.add("error");
                    assistantBubble.textContent = msg || "Error.";
                    setChatHint("Error received from server.");
                },
            },
            { signal: ac.signal }
        );

        // If aborted, treat as normal
        if (ac.signal.aborted) {
            assistantBubble.classList.add("error");
            assistantBubble.textContent = built ? `${built}\n\n[aborted]` : "[aborted]";
            setChatHint("Aborted.");
        }
    } catch (err) {
        if (ac.signal.aborted) {
            // handled above; just exit cleanly
            setChatHint("Aborted.");
        } else {
            assistantBubble.classList.add("error");
            assistantBubble.textContent = err?.message || "Request failed.";
            setChatHint("Request failed.");
        }
    } finally {
        setStreamingUI(false);
        setStatus("Idle");
        activeAbortController = null;
    }
});

// Initial state
setStatus("Idle");