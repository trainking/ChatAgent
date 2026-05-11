import * as esbuild from 'esbuild'

await esbuild.build({
  entryPoints: ['src/index.ts'],
  bundle: true,
  minify: true,
  format: 'iife',
  globalName: 'ChatAgentWidget',
  outfile: 'dist/widget.js',
  target: 'es2015',
})

console.log('✅ Widget built: dist/widget.js')
