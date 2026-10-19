export async function copyText(text: string, container: HTMLElement = document.body) {
  if (window.isSecureContext && navigator.clipboard) return navigator.clipboard.writeText(text)

  // The Clipboard API only exists on HTTPS/localhost; plain-HTTP LAN installs (CasaOS) need execCommand.
  const area = document.createElement("textarea")
  area.value = text
  area.readOnly = true
  area.style.position = "fixed"
  area.style.opacity = "0"
  container.appendChild(area)
  area.select()
  const ok = document.execCommand("copy")
  area.remove()
  if (!ok) throw new Error("copy failed")
}
