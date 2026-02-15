<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let files: any[] = [];
  export let disabled: boolean = false;

  let dragOver = false;

  const dispatch = createEventDispatcher();
</script>

<div class="file-list card">
  <div class="header">
    <h3>Files ({files.length})</h3>
    <div class="actions">
      <button class="primary" on:click={() => dispatch('browse')} {disabled}>
        + Add Files
      </button>
      {#if files.length > 0}
        <button class="secondary" on:click={() => dispatch('clear')} {disabled}>
          Clear All
        </button>
      {/if}
    </div>
  </div>

  {#if files.length === 0}
    <div class="empty" class:drag-over={dragOver}
         on:dragover|preventDefault={() => dragOver = true}
         on:dragleave={() => dragOver = false}
         on:drop|preventDefault={() => dragOver = false}
    >
      <p>{dragOver ? 'Drop files here' : 'No files added yet'}</p>
      <p class="hint">{dragOver ? '' : 'Click "Add Files" or drag & drop video/audio files'}</p>
    </div>
  {:else}
    <div class="list">
      {#each files as file (file.id)}
        <div class="file-item">
          <div class="file-info">
            <span class="file-name" title={file.path}>{file.name}</span>
            <span class="file-size">{file.sizeMb} MB</span>
          </div>
          <div class="file-right">
            <span class="badge {file.status}">{file.status}</span>
            {#if file.status === 'processing' && file.progress > 0}
              <span class="progress-text">{file.progress}%</span>
            {/if}
            {#if file.status === 'error' && file.error}
              <span class="error-text" title={file.error}>Error</span>
            {/if}
            {#if file.status === 'pending'}
              <button
                class="remove-btn"
                on:click={() => dispatch('remove', file.id)}
                {disabled}
                title="Remove"
              >
                x
              </button>
            {/if}
          </div>
        </div>
        {#if file.status === 'processing'}
          <div class="progress-bar-wrap">
            <div class="progress-bar" style="width: {file.progress}%"></div>
          </div>
        {/if}
        {#if file.status === 'done' && file.outputPath}
          <div class="output-path" title={file.outputPath}>{file.outputPath}</div>
        {/if}
      {/each}
    </div>
  {/if}
</div>

<style>
  .file-list {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  h3 {
    font-size: 14px;
    font-weight: 600;
  }

  .actions {
    display: flex;
    gap: 6px;
  }

  .empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    color: var(--text-muted);
    gap: 4px;
  }

  .hint {
    font-size: 12px;
    opacity: 0.7;
  }

  .list {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .file-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 8px;
    border-radius: 4px;
    background: var(--bg);
  }

  .file-info {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
  }

  .file-name {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
  }

  .file-size {
    color: var(--text-muted);
    font-size: 11px;
    white-space: nowrap;
  }

  .file-right {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-left: 8px;
  }

  .progress-text {
    font-size: 12px;
    color: var(--accent);
    font-weight: 600;
    min-width: 32px;
    text-align: right;
  }

  .error-text {
    font-size: 11px;
    color: var(--error);
    cursor: help;
  }

  .remove-btn {
    background: none;
    color: var(--text-muted);
    font-size: 14px;
    padding: 2px 6px;
    line-height: 1;
    border-radius: 4px;
  }

  .remove-btn:hover {
    background: var(--bg-hover);
    color: var(--error);
  }

  .progress-bar-wrap {
    height: 3px;
    background: var(--bg);
    border-radius: 2px;
    overflow: hidden;
    margin-top: -2px;
  }

  .progress-bar {
    height: 100%;
    background: var(--accent);
    transition: width 0.3s ease;
    border-radius: 2px;
  }

  .drag-over {
    border: 2px dashed var(--accent);
    background: rgba(59, 130, 246, 0.05);
    border-radius: 8px;
  }

  .output-path {
    font-size: 11px;
    color: var(--text-muted);
    padding: 2px 8px 4px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    opacity: 0.7;
  }
</style>
