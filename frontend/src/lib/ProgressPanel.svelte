<script lang="ts">
  export let modelDownloading: boolean = false;
  export let modelProgress: { percent: number; downloaded: string; total: string } | null = null;
  export let ffmpegDownloading: boolean = false;
  export let ffmpegProgress: { percent: number; downloaded: string; total: string } | null = null;
  export let modelLoading: boolean = false;
  export let statusMessage: string = '';
</script>

{#if modelDownloading || ffmpegDownloading || modelLoading || statusMessage}
  <div class="panel card">
    {#if modelDownloading && modelProgress}
      <div class="download-row">
        <span>Downloading model...</span>
        <span class="dl-stats">{modelProgress.downloaded} / {modelProgress.total} MB</span>
        <span class="dl-pct">{modelProgress.percent}%</span>
      </div>
      <div class="progress-bar-wrap">
        <div class="progress-bar" style="width: {modelProgress.percent}%"></div>
      </div>
    {:else if ffmpegDownloading && ffmpegProgress}
      <div class="download-row">
        <span>Downloading FFmpeg...</span>
        <span class="dl-stats">{ffmpegProgress.downloaded} / {ffmpegProgress.total} MB</span>
        <span class="dl-pct">{ffmpegProgress.percent}%</span>
      </div>
      <div class="progress-bar-wrap">
        <div class="progress-bar" style="width: {ffmpegProgress.percent}%"></div>
      </div>
    {:else if modelLoading}
      <div class="loading-row">
        <div class="spinner"></div>
        <span>Loading model into memory...</span>
      </div>
    {:else if statusMessage}
      <div class="status-row">
        <span>{statusMessage}</span>
      </div>
    {/if}
  </div>
{/if}

<style>
  .panel {
    flex-shrink: 0;
  }

  .download-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    margin-bottom: 6px;
  }

  .dl-stats {
    margin-left: auto;
    color: var(--text-muted);
    font-size: 12px;
  }

  .dl-pct {
    color: var(--accent);
    font-weight: 600;
    font-size: 13px;
    min-width: 36px;
    text-align: right;
  }

  .progress-bar-wrap {
    height: 6px;
    background: var(--bg);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background: var(--accent);
    transition: width 0.3s ease;
    border-radius: 3px;
  }

  .loading-row {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    color: var(--text-muted);
  }

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid var(--bg-hover);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .status-row {
    font-size: 13px;
    color: var(--text-muted);
  }
</style>
