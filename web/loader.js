// Downloads and compiles main.wasm behind a progress bar, then waits for a
// tap on the play button before the game starts. Starting from a tap lets the
// page go fullscreen and play the game's music and sounds, which browsers only
// allow in response to a user gesture.
'use strict';

(() => {
  const viewport =
    'width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no, viewport-fit=cover';

  const start = document.getElementById('start');
  const bar = document.getElementById('bar');
  const play = document.getElementById('play');
  const message = document.getElementById('message');
  const details = document.getElementById('details');

  // `make web` writes the byte size of main.wasm into the page. Under
  // wasmserve the placeholder stays, so the size comes from the response.
  const sizeHint = Number(document.documentElement.dataset.wasmSize);

  // Pinch zoom is off in the viewport tag, but Safari still zooms unless its
  // gesture events are cancelled. A long press must not open a menu either.
  document.addEventListener('gesturestart', (e) => e.preventDefault());
  document.addEventListener('contextmenu', (e) => e.preventDefault());

  function showProgress(fraction) {
    const percent = Math.min(100, Math.round(fraction * 100));
    bar.firstElementChild.style.width = percent + '%';
    bar.setAttribute('aria-valuenow', String(percent));
  }

  // fail replaces the loading screen with a message. When the game can be
  // retried, a tap anywhere reloads the page.
  function fail(text, detail, retry) {
    start.hidden = false;
    bar.hidden = true;
    play.hidden = true;
    message.textContent = text;
    if (detail) {
      details.textContent = detail;
      details.hidden = false;
    }
    if (retry) {
      start.classList.add('stopped');
      start.addEventListener('click', () => location.reload(), { once: true });
    }
  }

  // fetchWithProgress fetches a wasm file and moves the progress bar as it
  // arrives. The response it returns is always labelled application/wasm, so
  // streaming compilation works whatever type the server sent.
  async function fetchWithProgress(url) {
    const resp = await fetch(url);
    if (!resp.ok) {
      const err = new Error(`${url}: ${resp.status} ${resp.statusText}`);
      err.detail = await resp.text().catch(() => '');
      throw err;
    }

    let total = sizeHint > 0 ? sizeHint : 0;
    // The Content-Length of a compressed response counts compressed bytes,
    // while the reader below counts bytes after decompression.
    if (!total && !resp.headers.has('Content-Encoding')) {
      total = Number(resp.headers.get('Content-Length')) || 0;
    }

    const headers = { 'Content-Type': 'application/wasm' };
    if (!resp.body) {
      return new Response(await resp.arrayBuffer(), { headers });
    }
    if (!total) {
      bar.classList.add('waiting');
    }

    const reader = resp.body.getReader();
    let loaded = 0;
    const body = new ReadableStream({
      async pull(controller) {
        const { done, value } = await reader.read();
        if (done) {
          controller.close();
          return;
        }
        loaded += value.byteLength;
        if (total) {
          showProgress(loaded / total);
        }
        controller.enqueue(value);
      },
      cancel(reason) {
        return reader.cancel(reason);
      },
    });
    return new Response(body, { headers });
  }

  async function instantiate(resp, importObject) {
    if (WebAssembly.instantiateStreaming) {
      return WebAssembly.instantiateStreaming(resp, importObject);
    }
    return WebAssembly.instantiate(await resp.arrayBuffer(), importObject);
  }

  // Tablets and phones get the whole screen. A desktop browser keeps its
  // window, which suits development next to an editor.
  function enterFullscreen() {
    if (!window.matchMedia('(pointer: coarse)').matches) {
      return;
    }
    const root = document.documentElement;
    const request = root.requestFullscreen || root.webkitRequestFullscreen;
    if (!request) {
      return;
    }
    try {
      const pending = request.call(root);
      if (pending && pending.catch) {
        pending.catch(() => {});
      }
    } catch (e) {
      // Fullscreen is a nicety. The game runs in the page without it.
    }
  }

  // Ebitengine adds its own viewport tag when the game starts, which would
  // drop viewport-fit and the zoom limits. Keep every viewport tag on ours.
  function keepViewport() {
    const apply = () => {
      for (const meta of document.querySelectorAll('meta[name="viewport"]')) {
        if (meta.content !== viewport) {
          meta.content = viewport;
        }
      }
    };
    new MutationObserver(apply).observe(document.head, { childList: true });
  }

  function run(go, instance) {
    play.disabled = true;
    enterFullscreen();
    keepViewport();
    start.hidden = true;
    const stopped = (err) => {
      if (err) {
        console.error(err);
      }
      fail('The game stopped. Tap to start again.', '', true);
    };
    go.run(instance).then(() => stopped(), stopped);
  }

  async function load() {
    if (typeof WebAssembly !== 'object') {
      fail('This browser cannot run the game. Please try a newer one.', '', false);
      return;
    }
    if (typeof Go !== 'function') {
      fail('The game did not load. Tap to try again.', 'wasm_exec.js is missing.', true);
      return;
    }

    const go = new Go();
    let instance;
    try {
      const resp = await fetchWithProgress('main.wasm');
      instance = (await instantiate(resp, go.importObject)).instance;
    } catch (err) {
      console.error(err);
      fail('The game did not load. Tap to try again.', err.detail || String(err), true);
      return;
    }

    showProgress(1);
    bar.hidden = true;
    play.hidden = false;
    play.focus({ preventScroll: true });
    play.addEventListener('click', () => run(go, instance), { once: true });
  }

  // registerServiceWorker lets the game start offline after one online visit.
  // Workers need https or localhost, so on a plain http LAN address there is
  // none. The game runs the same without one, so a failure is only logged.
  function registerServiceWorker() {
    if (!('serviceWorker' in navigator)) {
      return;
    }
    window.addEventListener('load', () => {
      navigator.serviceWorker.register('sw.js').catch((err) => {
        console.warn('Service worker not registered:', err);
      });
    });
  }

  registerServiceWorker();
  load();
})();
