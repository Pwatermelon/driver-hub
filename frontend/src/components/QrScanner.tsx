import { useEffect, useRef, useState } from 'react'
import jsQR from 'jsqr'

type Props = {
  onClose: () => void
  onScan: (hubId: string) => void
}

function extractHub(raw: string): string | null {
  const value = raw.trim()
  if (!value) return null
  try {
    const url = new URL(value)
    const hub = url.searchParams.get('hub')
    if (hub?.trim()) return hub.trim()
  } catch {
    /* not a URL */
  }
  const match = value.match(/DH-[A-Z0-9]+/i)
  if (match) return match[0].toUpperCase()
  if (/^[A-Z0-9-]{6,}$/i.test(value)) return value.toUpperCase()
  return null
}

export function QrScanner({ onClose, onScan }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [error, setError] = useState('')
  const [ready, setReady] = useState(false)
  const doneRef = useRef(false)

  useEffect(() => {
    let stream: MediaStream | null = null
    let raf = 0
    let cancelled = false
    const video = videoRef.current

    async function start() {
      try {
        stream = await navigator.mediaDevices.getUserMedia({
          audio: false,
          video: { facingMode: { ideal: 'environment' }, width: { ideal: 1280 }, height: { ideal: 720 } },
        })
        if (cancelled) {
          stream.getTracks().forEach((t) => t.stop())
          return
        }
        if (!video) return
        video.srcObject = stream
        await video.play()
        setReady(true)

        const canvas = canvasRef.current
        const ctx = canvas?.getContext('2d', { willReadFrequently: true })
        if (!canvas || !ctx) return

        const tick = () => {
          if (cancelled || doneRef.current) return
          if (video.readyState >= 2) {
            const w = video.videoWidth
            const h = video.videoHeight
            if (w && h) {
              canvas.width = w
              canvas.height = h
              ctx.drawImage(video, 0, 0, w, h)
              const image = ctx.getImageData(0, 0, w, h)
              const code = jsQR(image.data, w, h, { inversionAttempts: 'dontInvert' })
              if (code?.data) {
                const hub = extractHub(code.data)
                if (hub) {
                  doneRef.current = true
                  onScan(hub)
                  onClose()
                  return
                }
              }
            }
          }
          raf = requestAnimationFrame(tick)
        }
        raf = requestAnimationFrame(tick)
      } catch {
        if (!cancelled) {
          setError('Не удалось открыть камеру. Разрешите доступ или введите Hub ID вручную.')
        }
      }
    }

    void start()

    return () => {
      cancelled = true
      cancelAnimationFrame(raf)
      stream?.getTracks().forEach((t) => t.stop())
      if (video) video.srcObject = null
    }
  }, [onClose, onScan])

  return (
    <div className="qr-modal" role="dialog" aria-modal="true" aria-label="Сканирование QR водителя">
      <div className="qr-modal-panel">
        <div className="qr-modal-head">
          <div>
            <h2>Сканировать QR водителя</h2>
            <p className="sub">Наведите камеру на QR с телефона водителя</p>
          </div>
          <button type="button" className="btn" onClick={onClose}>
            Закрыть
          </button>
        </div>
        <div className="qr-viewport">
          <video ref={videoRef} playsInline muted />
          <div className="qr-frame-guide" aria-hidden />
          {!ready && !error && <div className="qr-overlay-msg">Открываем камеру…</div>}
          {error && <div className="qr-overlay-msg">{error}</div>}
        </div>
        <canvas ref={canvasRef} className="qr-canvas" />
        <p className="micro qr-hint">QR должен содержать Hub ID или ссылку вида /company?hub=DH-…</p>
      </div>
    </div>
  )
}
