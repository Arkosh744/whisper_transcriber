<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let languages: { code: string; name: string }[] = [];
  export let language: string = 'auto';
  export let outputFormat: string = 'srt';
  export let isRunning: boolean = false;
  export let hasFiles: boolean = false;
  export let modelReady: boolean = false;
  export let ffmpegReady: boolean = false;

  const dispatch = createEventDispatcher();

  const formats = [
    { value: 'srt', label: 'SRT (subtitles)' },
    { value: 'txt', label: 'TXT (timestamps)' },
    { value: 'json', label: 'JSON (structured)' },
    { value: 'md', label: 'Markdown' },
  ];
</script>

<div class="controls card">
  <div class="row">
    <div class="field">
      <label for="lang">Language</label>
      <select id="lang" bind:value={language} disabled={isRunning}>
        {#each languages as lang}
          <option value={lang.code}>{lang.name}</option>
        {/each}
      </select>
    </div>
    <div class="field">
      <label for="format">Output</label>
      <select id="format" bind:value={outputFormat} disabled={isRunning}>
        {#each formats as fmt}
          <option value={fmt.value}>{fmt.label}</option>
        {/each}
      </select>
    </div>
    <div class="buttons">
      {#if !isRunning}
        <button
          class="primary start-btn"
          disabled={!hasFiles || !modelReady || !ffmpegReady}
          on:click={() => dispatch('start')}
        >
          Start Transcription
        </button>
      {:else}
        <button class="danger" on:click={() => dispatch('cancel')}>
          Cancel
        </button>
      {/if}
    </div>
  </div>

  {#if !modelReady}
    <div class="model-notice">
      <span class="warn-icon">!</span>
      Model not found.
      <button class="link-btn" on:click={() => dispatch('download-model')}>
        Download model (~574 MB)
      </button>
    </div>
  {/if}

  {#if !ffmpegReady}
    <div class="model-notice">
      <span class="warn-icon">!</span>
      FFmpeg not found.
      <button class="link-btn" on:click={() => dispatch('download-ffmpeg')}>
        Download FFmpeg (~200 MB)
      </button>
    </div>
  {/if}
</div>

<style>
  .controls {
    flex-shrink: 0;
  }

  .row {
    display: flex;
    gap: 12px;
    align-items: flex-end;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    font-weight: 600;
    letter-spacing: 0.5px;
  }

  select {
    min-width: 150px;
  }

  .buttons {
    margin-left: auto;
    display: flex;
    gap: 8px;
  }

  .start-btn {
    padding: 8px 24px;
    font-weight: 600;
  }

  .model-notice {
    margin-top: 8px;
    padding: 6px 10px;
    background: #451a03;
    border: 1px solid #92400e;
    border-radius: 6px;
    font-size: 12px;
    color: var(--warning);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .warn-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--warning);
    color: #000;
    font-weight: 700;
    font-size: 12px;
    flex-shrink: 0;
  }

  .link-btn {
    background: none;
    color: var(--accent);
    text-decoration: underline;
    padding: 0;
    font-size: 12px;
  }

  .link-btn:hover {
    color: var(--accent-hover);
  }
</style>
