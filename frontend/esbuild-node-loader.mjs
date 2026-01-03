import esbuild from 'esbuild'
import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

export async function resolve(specifier, context, defaultResolve) {
  if (specifier.startsWith('.') || specifier.startsWith('/') || specifier.startsWith('../')) {
    const parentPath = context.parentURL ? path.dirname(fileURLToPath(context.parentURL)) : process.cwd()
    const candidate = path.resolve(parentPath, specifier)
    const tsCandidate = `${candidate}.ts`
    try {
      await fs.access(tsCandidate)
      return defaultResolve(pathToFileURL(tsCandidate).href, context, defaultResolve)
    } catch {
      // fallback
    }
  }
  return defaultResolve(specifier, context, defaultResolve)
}

export async function load(url, context, defaultLoad) {
  if (url.endsWith('.ts')) {
    const source = await fs.readFile(new URL(url))
    const { code } = await esbuild.transform(source.toString(), {
      loader: 'ts',
      format: 'esm',
      target: 'es2022',
      sourcemap: 'inline'
    })
    return { format: 'module', source: code, shortCircuit: true }
  }
  return defaultLoad(url, context, defaultLoad)
}
