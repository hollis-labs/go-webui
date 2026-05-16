package webui

// DefaultPlaceholder is the HTML served when no built SPA is present in
// the configured fs.FS. Override it per host via Config.Placeholder.
const DefaultPlaceholder = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Web UI not built</title>
  <style>
    body { margin: 0; background: #0a0a0a; color: #e5e7eb; font-family: ui-sans-serif, system-ui, -apple-system, sans-serif; }
    main { display: flex; align-items: center; justify-content: center; min-height: 100vh; padding: 24px; text-align: center; }
    h1 { margin: 0; font-size: 1.5rem; font-weight: 600; }
    p { margin: 0.75rem 0 0; color: #9ca3af; }
    code { color: #f9fafb; }
  </style>
</head>
<body>
  <main>
    <div>
      <h1>Web UI not built</h1>
      <p>Run the UI build step and rebuild the binary.</p>
    </div>
  </main>
</body>
</html>`
