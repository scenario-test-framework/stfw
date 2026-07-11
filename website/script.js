const translations = {
  ja: {
    pageTitle: "stfw — Scenario Test Framework",
    description: "stfwは、業務日をまたぐシナリオテストをGo製シングルバイナリひとつで実行するテストフレームワークです。",
    ogDescription: "複雑なシナリオを、確信をもってリリース。",
    ogLocale: "ja_JP",
    docs: "https://github.com/scenario-test-framework/stfw/blob/master/README.ja.md",
    docsQuickstart: "https://github.com/scenario-test-framework/stfw/blob/master/README.ja.md#quick-start",
    skipToContent: "本文へ移動",
    navFeatures: "特徴",
    navPlugins: "プラグイン",
    navDocs: "ドキュメント",
    viewDocs: "ドキュメントを見る",
    getStarted: "はじめる",
    eyebrow: "シナリオテストフレームワーク",
    heroLine1: "複雑なシナリオを",
    heroLine2: "確信をもってリリース",
    heroLeadLine1: "業務日をまたぐテストを、",
    heroLeadLine2: "Go製シングルバイナリひとつで。",
    viewGithub: "GitHubで見る",
    businessDate: "業務日",
    singleBinary: "シングルバイナリ",
    failFast: "フェイルファスト実行",
    otlpNative: "OTLPネイティブ",
    featuresTitle: "実システムのシナリオテストのために",
    conventionTitle: "設定より規約",
    conventionDesc: "シンプルなディレクトリ規約で、シナリオを整理。",
    visibleTitle: "すべてのステップを可視化",
    visibleDesc: "構造化ログとトレースで、実行状況をリアルタイムに把握。",
    composeTitle: "テスト全体を自在に構成",
    composeDesc: "プラグインでシステム、サービス、ツールを連携。",
    morePlugins: "＋ その他のプラグイン",
    workflowTitle: "明快で一貫したワークフロー",
    arrangeDesc: "データと前提条件を準備",
    actDesc: "システム上の操作を実行",
    collectDesc: "結果と観測データを収集",
    assertDesc: "結果と業務ルールを検証",
    executionReport: "実行レポート",
    duration: "所要時間",
    success: "成功",
    reportLine1: "実行から",
    reportLine2: "根本原因まで",
    reportDesc: "レポート、タイムライン、トレースで、失敗から修正までを高速化。",
    businessDateOverview: "業務日ごとの概要",
    stepStatus: "ステップ単位の状態",
    durationsTimelines: "所要時間とタイムライン",
    logsTracesMetrics: "ログ、トレース、メトリクス",
    quickLine1: "1分以内に",
    quickLine2: "スタート。",
    finalLine1: "すべての業務日を、",
    finalLine2: "テスト可能に。",
    copy: "コピー",
    copied: "コピー済み",
    copyFallback: "選択してコピー",
    homeLabel: "stfw ホーム",
    mainNavLabel: "メインナビゲーション",
    scenarioMapLabel: "シナリオが業務日ごとに実行され、成功に至る流れ",
    pipelineLabel: "テストの4フェーズ",
    proofLabel: "主要な特徴",
    directoryExampleLabel: "シナリオのディレクトリ例",
    pluginListLabel: "対応プラグイン",
    pluginRailLabel: "ワークフローを構成するプラグイン",
    reportExampleLabel: "daily-balanceシナリオの実行レポート例",
    durationTimelineLabel: "実行時間のタイムライン",
    quickCommandsLabel: "クイックスタートコマンド",
    copyCommandLabel: "コマンドをコピー",
    platformListLabel: "対応プラットフォーム",
    backToTopLabel: "ページ上部へ",
    footerNavLabel: "フッターナビゲーション",
    toggleLabel: "Switch to English",
    toggleText: "EN",
  },
  en: {
    pageTitle: "stfw — Scenario Test Framework",
    description: "stfw runs cross-business-date scenario tests from a predictable directory convention with a single Go binary.",
    ogDescription: "Ship complex scenarios with confidence.",
    ogLocale: "en_US",
    docs: "https://github.com/scenario-test-framework/stfw/blob/master/README.md",
    docsQuickstart: "https://github.com/scenario-test-framework/stfw/blob/master/README.md#quick-start",
    skipToContent: "Skip to content",
    navFeatures: "Features",
    navPlugins: "Plugins",
    navDocs: "Docs",
    viewDocs: "View Docs",
    getStarted: "Get Started",
    eyebrow: "SCENARIO TEST FRAMEWORK",
    heroLine1: "Ship complex scenarios",
    heroLine2: "with confidence.",
    heroLeadLine1: "Cross-business-date testing.",
    heroLeadLine2: "One Go binary. Zero workflow engine.",
    viewGithub: "View on GitHub",
    businessDate: "BUSINESS DATE",
    singleBinary: "Single binary",
    failFast: "Fail-fast execution",
    otlpNative: "OTLP native",
    featuresTitle: "Built for real system scenarios",
    conventionTitle: "Convention over configuration",
    conventionDesc: "Organize scenarios with a simple, predictable directory structure.",
    visibleTitle: "Every step is visible",
    visibleDesc: "Follow execution in real time with structured logs and traces.",
    composeTitle: "Compose the full test",
    composeDesc: "Use plugins to orchestrate systems, services, and tooling.",
    morePlugins: "+ More plugins",
    workflowTitle: "A clear, consistent workflow",
    arrangeDesc: "Set up data and preconditions",
    actDesc: "Execute actions in the system",
    collectDesc: "Gather results and observability",
    assertDesc: "Verify outcomes and business rules",
    executionReport: "Execution report",
    duration: "Duration",
    success: "Success",
    reportLine1: "From run to",
    reportLine2: "root cause.",
    reportDesc: "Rich reports, timelines, and traces help you move from failure to fix faster.",
    businessDateOverview: "Business-date overview",
    stepStatus: "Step-level status",
    durationsTimelines: "Durations & timelines",
    logsTracesMetrics: "Logs, traces, and metrics",
    quickLine1: "Start in under",
    quickLine2: "a minute.",
    finalLine1: "Make every business date",
    finalLine2: "testable.",
    copy: "Copy",
    copied: "Copied",
    copyFallback: "Select to copy",
    homeLabel: "stfw home",
    mainNavLabel: "Main navigation",
    scenarioMapLabel: "Scenarios flowing through business dates to a successful result",
    pipelineLabel: "Four test phases",
    proofLabel: "Key features",
    directoryExampleLabel: "Example scenario directory",
    pluginListLabel: "Supported plugins",
    pluginRailLabel: "Plugins used to compose the workflow",
    reportExampleLabel: "Example execution report for the daily-balance scenario",
    durationTimelineLabel: "Execution duration timeline",
    quickCommandsLabel: "Quick-start commands",
    copyCommandLabel: "Copy commands",
    platformListLabel: "Supported platforms",
    backToTopLabel: "Back to top",
    footerNavLabel: "Footer navigation",
    toggleLabel: "日本語に切り替える",
    toggleText: "日本語",
  },
};

const languageToggle = document.querySelector("[data-language-toggle]");
const copyButton = document.querySelector("[data-copy-target]");
let currentLanguage = "ja";

function browserLanguage() {
  const preferred = navigator.languages?.[0] || navigator.language || "en";
  return preferred.toLowerCase().startsWith("ja") ? "ja" : "en";
}

function initialLanguage() {
  const queryLanguage = new URLSearchParams(window.location.search).get("lang");
  if (queryLanguage === "ja" || queryLanguage === "en") return queryLanguage;

  const storedLanguage = window.localStorage.getItem("stfw-language");
  if (storedLanguage === "ja" || storedLanguage === "en") return storedLanguage;

  return browserLanguage();
}

function setLanguage(language, persist = false) {
  const dictionary = translations[language] || translations.en;
  currentLanguage = language in translations ? language : "en";
  document.documentElement.lang = currentLanguage;

  document.querySelectorAll("[data-i18n]").forEach((element) => {
    const value = dictionary[element.dataset.i18n];
    if (value) element.textContent = value;
  });

  document.querySelectorAll("[data-i18n-aria]").forEach((element) => {
    const value = dictionary[element.dataset.i18nAria];
    if (value) element.setAttribute("aria-label", value);
  });

  document.querySelectorAll("[data-i18n-href]").forEach((element) => {
    const value = dictionary[element.dataset.i18nHref];
    if (value) element.setAttribute("href", value);
  });

  document.title = dictionary.pageTitle;
  document.querySelector('meta[name="description"]')?.setAttribute("content", dictionary.description);
  document.querySelector('meta[property="og:description"]')?.setAttribute("content", dictionary.ogDescription);
  document.querySelector('meta[property="og:locale"]')?.setAttribute("content", dictionary.ogLocale);

  if (languageToggle) {
    languageToggle.textContent = dictionary.toggleText;
    languageToggle.setAttribute("aria-label", dictionary.toggleLabel);
  }

  if (copyButton) copyButton.textContent = dictionary.copy;
  if (persist) window.localStorage.setItem("stfw-language", currentLanguage);
}

setLanguage(initialLanguage());

languageToggle?.addEventListener("click", () => {
  setLanguage(currentLanguage === "ja" ? "en" : "ja", true);
});

copyButton?.addEventListener("click", async () => {
  const target = document.getElementById(copyButton.dataset.copyTarget);
  if (!target) return;

  const commands = target.textContent
    .split("\n")
    .map((line) => line.replace(/^\$\s*/, ""))
    .join("\n")
    .trim();

  try {
    await navigator.clipboard.writeText(commands);
    copyButton.textContent = translations[currentLanguage].copied;
    window.setTimeout(() => {
      copyButton.textContent = translations[currentLanguage].copy;
    }, 1800);
  } catch {
    copyButton.textContent = translations[currentLanguage].copyFallback;
  }
});
