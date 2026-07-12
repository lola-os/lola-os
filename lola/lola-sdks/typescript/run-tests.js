/**
 * Minimal test runner: imports every compiled *.test.js file under
 * dist-test/test/, and calls every exported `test_*` function (running
 * `teardown()` after a file's tests if exported). No external test
 * framework dependency, since this environment has no network access to
 * install one — written to be drop-in replaceable by Jest/Mocha/Vitest
 * later (the test bodies use plain Node `assert`, which all of those
 * support natively).
 */
const fs = require("fs");
const path = require("path");

const testDir = path.join(__dirname, "dist-test", "test");
const files = fs.readdirSync(testDir).filter((f) => f.endsWith(".test.js"));

let total = 0;
let failed = 0;

async function run() {
  for (const file of files) {
    const mod = require(path.join(testDir, file));
    for (const [name, fn] of Object.entries(mod)) {
      if (!name.startsWith("test_") || typeof fn !== "function") continue;
      total += 1;
      try {
        await fn();
        console.log(`PASS ${file}::${name}`);
      } catch (err) {
        failed += 1;
        console.log(`FAIL ${file}::${name}: ${err && err.stack ? err.stack : err}`);
      }
    }
    if (typeof mod.teardown === "function") {
      await mod.teardown();
    }
  }
  console.log(`\n${total - failed}/${total} passed`);
  process.exit(failed > 0 ? 1 : 0);
}

run();
