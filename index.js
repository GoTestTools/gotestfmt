const core = require('@actions/core');
const child_process = require("child_process");
const os = require("os");

try {
    const targetDir = process.env.GITHUB_WORKSPACE
    const gotestfmt = path.join(process.cwd(), "cmd", "gotestfmt")
    const proc = child_process.spawn("go run " + gotestfmt, {
        cwd: targetDir,
    })
    proc.on("data", (data) => {
        os.stdout.write(data)
    })
    proc.on("close", (code) => {
        process.exit(code)
    })
} catch (error) {
    core.setFailed(error.message);
}