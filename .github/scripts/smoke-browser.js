// smoke-browser.js は stfw:full の Chrome が k6 ブラウザモードから起動し、ページを読み込んで DOM を操作できるかを確かめる
// (smoke-image.sh から呼ぶ。ネットワークなしで動くよう setContent を使う。日本語フォントの有無は smoke-image.sh が見る)。
import { browser } from 'k6/browser';
import { check } from 'k6';

export const options = {
  scenarios: { ui: { executor: 'shared-iterations', options: { browser: { type: 'chromium' } } } },
  thresholds: { checks: ['rate==1'] },
};

export default async function () {
  const page = await browser.newPage();
  try {
    await page.setContent('<html><body><h1 id="t">日本語の表示</h1></body></html>');
    const text = await page.locator('#t').textContent();
    check(text, { 'heading rendered': (t) => t === '日本語の表示' });
  } finally {
    await page.close();
  }
}
