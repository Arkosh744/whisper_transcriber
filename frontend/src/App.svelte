<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff, OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime';
  import {
    BrowseFiles,
    AddFiles,
    ClearFiles,
    RemoveFile,
    GetLanguages,
    IsModelAvailable,
    DownloadModel,
    IsFFmpegAvailable,
    DownloadFFmpeg,
    StartTranscription,
    CancelTranscription,
    CancelDownload,
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
  let cancelling = false;

  // Progress panel state
  let modelDownloading = false;
  let modelProgress: { percent: number; downloaded: string; total: string } | null = null;
  let ffmpegDownloading = false;
  let ffmpegProgress: { percent: number; downloaded: string; total: string } | null = null;
  let modelLoading = false;
  let statusMessage = '';

  $: if (typeof window !== 'undefined') localStorage.setItem('wt:language', language);
  $: if (typeof window !== 'undefined') localStorage.setItem('wt:outputFormat', outputFormat);

  // Cleanup handles
  let cleanups: (() => void)[] = [];

  function on(event: string, cb: (...args: any[]) => void) {
    EventsOn(event, cb);
    cleanups.push(() => EventsOff(event));
  }

  onMount(async () => {
    // Restore saved settings
    const savedLang = localStorage.getItem('wt:language');
    if (savedLang) language = savedLang;
    const savedFormat = localStorage.getItem('wt:outputFormat');
    if (savedFormat) outputFormat = savedFormat;

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
      statusMessage = cancelling ? 'Cancelled' : 'Batch complete!';
      cancelling = false;
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
      files = files.map(f =>
        f.id === data.fileID ? { ...f, outputPath: data.outputPath } : f
      );
    });

    OnFileDrop(async (_x: number, _y: number, paths: string[]) => {
      try {
        const items = await AddFiles(paths);
        if (items && items.length > 0) {
          files = [...files, ...items];
        }
      } catch (e: any) {
        statusMessage = 'Error: ' + (e?.message || e);
      }
    }, true);
  });

  onDestroy(() => {
    cleanups.forEach(fn => fn());
    OnFileDropOff();
  });

  // Handlers
  async function handleBrowse() {
    try {
      const items = await BrowseFiles();
      if (items && items.length > 0) {
        files = [...files, ...items];
      }
    } catch (e) {
      statusMessage = 'Error: ' + e;
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
    cancelling = true;
    statusMessage = 'Cancelling...';
  }

  function handleCancelDownload() {
    CancelDownload();
    modelDownloading = false;
    ffmpegDownloading = false;
    modelProgress = null;
    ffmpegProgress = null;
    statusMessage = 'Download cancelled';
    setTimeout(() => { statusMessage = ''; }, 2000);
  }

  function handleDownloadModel() {
    try {
      DownloadModel();
      modelDownloading = true;
    } catch (e: any) {
      statusMessage = 'Error: ' + (e?.message || e);
    }
  }

  function handleDownloadFFmpeg() {
    try {
      DownloadFFmpeg();
      ffmpegDownloading = true;
    } catch (e: any) {
      statusMessage = 'Error: ' + (e?.message || e);
    }
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
  on:cancel-download={handleCancelDownload}
/>

<Controls
  {languages}
  bind:language
  bind:outputFormat
  {isRunning}
  {cancelling}
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
