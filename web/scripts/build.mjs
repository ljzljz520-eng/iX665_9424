import { mkdir, cp } from "node:fs/promises";

await mkdir("dist", { recursive: true });
await cp("index.html", "dist/index.html");
