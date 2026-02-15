<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
  import {
    BrowseFiles,
    ClearFiles,
    RemoveFile,
    GetLanguages,
    IsModelAvailable,
    DownloadModel,
    IsFFmpegAvailable,
    DownloadFFmpeg,
    StartTranscription,
    CancelTranscription,
  } from '../wailsjs/go/main/App';
  import FileList from './lib/FileList.svelte';
  import Controls from './lib/Controls.svelte';
  import ProgressPanel from './lib/ProgressPanel.svelte';

  // State
  let files: any[] = [];
  let languages: { code: string; name: string }[] = [];
  let language = 'auto';
  let outputFormat = 'srt';
  let modelReady = false;
  let ffmpegReady = false;
  let isRunning = false;

  // Progress panel state
  let modelDownloading = false;
  let modelProgress: { percent: number; downloaded: string; total: string } | null = null;
  let ffmpegDownloading = false;
  let ffmpegProgress: { percent: number; downloaded: string; total: string } | null = null;
  let modelLoading = false;
  let statusMessage = '';

  // Cleanup handles
  let cleanups: (() => void)[] = [];

  function on(event: string, cb: (...args: any[]) => void) {
    EventsOn(event, cb);
    cleanups.push(() => EventsOff(event));
  }

  onMount(async () => {
    // Load initial data
    languages = await GetLanguages();
    modelReady = await IsModelAvailable();
    ffmpegReady = await IsFFmpegAvailable();

    // File status events
    on('file:status', (data: any) => {
      files = files.map(f =>
        f.id === data.fileID
          ? { ...f, status: data.status, progress: data.progress, error: data.error }
          : f
      );
    });

    // Transcription progress
    on('transcription:progress', (data: any) => {
      files = files.map(f =>
        f.id === data.fileID ? { ...f, progress: data.progress } : f
      );
    });

    // Batch complete
    on('batch:complete', () => {
      isRunning = false;
      statusMessage = 'Batch complete!';
      setTimeout(() => { statusMessage = ''; }, 3000);
    });

    // Model events
    on('model:loading', () => {
      modelLoading = true;
    });

    on('model:loaded', () => {
      modelLoading = false;
    });

    on('model:download:progress', (data: any) => {
      modelDownloading = true;
      modelProgress = data;
    });

    on('model:download:done', () => {
      modelDownloading = false;
      modelProgress = null;
      modelReady = true;
      statusMessage = 'Model downloaded!';
      setTimeout(() => { statusMessage = ''; }, 3000);
    });

    on('model:download:error', (errMsg: string) => {
      modelDownloading = false;
      modelProgress = null;
      statusMessage = 'Download error: ' + errMsg;
    });

    // FFmpeg events
    on('ffmpeg:download:progress', (data: any) => {
      ffmpegDownloading = true;
      ffmpegProgress = data;
    });

    on('ffmpeg:download:done', () => {
      ffmpegDownloading = false;
      ffmpegProgress = null;
      ffmpegReady = true;
      statusMessage = 'FFmpeg downloaded!';
      setTimeout(() => { statusMessage = ''; }, 3000);
    });

    on('ffmpeg:download:error', (errMsg: string) => {
      ffmpegDownloading = false;
      ffmpegProgress = null;
      statusMessage = 'FFmpeg download error: ' + errMsg;
    });

    on('transcription:complete', (data: any) => {
      // File transcribed successfully
    });
  });

  onDestroy(() => {
    cleanups.forEach(fn => fn());
  });

  // Handlers
  async function handleBrowse() {
    try {
      const items = await BrowseFiles();
      if (items && items.length > 0) {
        files = [...files, ...items];
      }
    } catch (e) {
      console.error('Browse error:', e);
    }
  }

  function handleClear() {
    ClearFiles();
    files = [];
  }

  function handleRemove(e: CustomEvent<string>) {
    RemoveFile(e.detail);
    files = files.filter(f => f.id !== e.detail);
  }

  async function handleStart() {
    if (files.length === 0 || !modelReady) return;
    isRunning = true;
    statusMessage = '';
    try {
      await StartTranscription({ language, outputFormat });
    } catch (e: any) {
      isRunning = false;
      statusMessage = 'Error: ' + (e?.message || e);
    }
  }

  function handleCancel() {
    CancelTranscription();
    isRunning = false;
    statusMessage = 'Cancelled';
    setTimeout(() => { statusMessage = ''; }, 2000);
  }

  function handleDownloadModel() {
    DownloadModel();
    modelDownloading = true;
  }

  function handleDownloadFFmpeg() {
    DownloadFFmpeg();
    ffmpegDownloading = true;
  }
</script>

<div class="app-header">
  <h1>Whisper Transcriber</h1>
</div>

<ProgressPanel
  {modelDownloading}
  {modelProgress}
  {ffmpegDownloading}
  {ffmpegProgress}
  {modelLoading}
  {statusMessage}
/>

<Controls
  {languages}
  bind:language
  bind:outputFormat
  {isRunning}
  hasFiles={files.length > 0}
  {modelReady}
  {ffmpegReady}
  on:start={handleStart}
  on:cancel={handleCancel}
  on:download-model={handleDownloadModel}
  on:download-ffmpeg={handleDownloadFFmpeg}
/>

<FileList
  {files}
  disabled={isRunning}
  on:browse={handleBrowse}
  on:clear={handleClear}
  on:remove={handleRemove}
/>

<style>
  .app-header {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  h1 {
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.3px;
  }
</style>
