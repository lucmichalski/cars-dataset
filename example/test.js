const playwright = require("playwright");

const browserArgs = {
  chromium: {
    // these arguments are required for chromium
    args: ["--disable-dev-shm-usage", "--no-sandbox", "--disable-setuid-sandbox"]
  }
};

const defaultViewport = {
  width: 1280,
  height: 720
};

(async () => {
  for (const browserType of ["firefox", "webkit", "chromium"]) {
    try {
      console.log(`browserType=${browserType}`);
      const args = browserArgs[browserType] || {};
      const browser = await playwright[browserType].launch(args);
      const context = await browser.newContext({
        viewport: defaultViewport
      });
      const page = await context.newPage();
      const url = "https://example.com";
      await page.goto(url);

      await page.screenshot({ path: `screenshots/example-${browserType}.png` });
      await browser.close();
    } catch(e) {
      console.error(e);
    }
  }
})();
