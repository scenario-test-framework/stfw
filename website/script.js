const copyButton = document.querySelector("[data-copy-target]");

if (copyButton) {
  copyButton.addEventListener("click", async () => {
    const target = document.getElementById(copyButton.dataset.copyTarget);
    if (!target) return;

    const commands = target.textContent
      .split("\n")
      .map((line) => line.replace(/^\$\s*/, ""))
      .join("\n")
      .trim();

    try {
      await navigator.clipboard.writeText(commands);
      copyButton.textContent = "コピー済み";
      window.setTimeout(() => {
        copyButton.textContent = "コピー";
      }, 1800);
    } catch {
      copyButton.textContent = "選択してコピー";
    }
  });
}
