import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const srcDir = path.resolve(__dirname, "../packages/cds/src");

function walk(dir) {
  const files = [];
  fs.readdirSync(dir, { withFileTypes: true }).forEach((entry) => {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) files.push(...walk(full));
    else if (entry.name.endsWith(".tsx") || entry.name.endsWith(".ts"))
      files.push(full);
  });
  return files;
}

const files = walk(srcDir);
let changed = 0;

files.forEach((filePath) => {
  let content = fs.readFileSync(filePath, "utf8");
  const original = content;

  const fileDir = path.dirname(filePath);
  const relToSrc = path.relative(fileDir, srcDir).replace(/\\/g, "/") || ".";

  // Replace all "@/" imports with relative paths
  content = content.replace(
    /(["'])@\/([^"']+)["']/g,
    (match, quote, importPath) => {
      return quote + relToSrc + "/" + importPath + quote;
    }
  );

  if (content !== original) {
    fs.writeFileSync(filePath, content, "utf8");
    changed++;
    console.log("Updated:", path.relative(process.cwd(), filePath));
  }
});

console.log(`\n${changed} files updated.`);