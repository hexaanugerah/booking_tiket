import { existsSync, statSync } from "node:fs";
import path from "node:path";
import { registerHooks } from "node:module";
import { fileURLToPath, pathToFileURL } from "node:url";

// Test runner menjalankan modul .ts apa adanya (type stripping bawaan Node).
// Hook ini menyelesaikan alias "@/..." dari tsconfig dan import relative tanpa
// ekstensi, karena ESM mewajibkan ekstensi sedangkan kode aplikasi tidak memakainya.

const projectRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const EXTENSIONS = ["", ".ts", ".tsx", "/index.ts", "/index.tsx"];

function candidateBases(specifier, parentURL) {
  if (specifier.startsWith("@/")) {
    return [path.resolve(projectRoot, specifier.slice(2))];
  }
  if (specifier.startsWith(".") && parentURL?.startsWith("file:")) {
    return [path.resolve(path.dirname(fileURLToPath(parentURL)), specifier)];
  }
  return [];
}

function resolveFile(base) {
  for (const suffix of EXTENSIONS) {
    const candidate = base + suffix;
    if (existsSync(candidate) && statSync(candidate).isFile()) {
      return candidate;
    }
  }
  return null;
}

registerHooks({
  resolve(specifier, context, nextResolve) {
    // Hanya untuk ESM: resolver CJS tidak menerima URL file:// dan sudah bisa
    // menyelesaikan dependency node_modules seperti biasa.
    if (!context.conditions?.includes("import")) {
      return nextResolve(specifier, context);
    }
    for (const base of candidateBases(specifier, context.parentURL)) {
      const file = resolveFile(base);
      if (file) return nextResolve(pathToFileURL(file).href, context);
    }
    return nextResolve(specifier, context);
  },
});
