export type ExtractedColor = { hex: string; ratio: number }

type Pixel = [number, number, number]

/**
 * 从照片里提取主色：缩放 → 取像素 → 中位切分配色量化 → 按占比排序。
 * 纯前端实现，照片不会离开浏览器。
 */
export async function extractPhotoPalette(file: File, count = 6): Promise<{ colors: ExtractedColor[]; preview: string }> {
  const bitmap = await loadBitmap(file)
  const maxEdge = 220
  const scale = Math.min(1, maxEdge / Math.max(bitmap.width, bitmap.height))
  const width = Math.max(1, Math.round(bitmap.width * scale))
  const height = Math.max(1, Math.round(bitmap.height * scale))
  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height
  const context = canvas.getContext('2d', { willReadFrequently: true })
  if (!context) throw new Error('当前浏览器无法读取图片像素')
  context.drawImage(bitmap, 0, 0, width, height)
  const { data } = context.getImageData(0, 0, width, height)

  const pixels: Pixel[] = []
  for (let index = 0; index < data.length; index += 4) {
    if (data[index + 3] < 200) continue
    pixels.push([data[index], data[index + 1], data[index + 2]])
  }
  if (pixels.length < 16) throw new Error('照片里的有效像素太少，换一张试试')
  if ('close' in bitmap) bitmap.close()

  const boxes = medianCut(pixels, Math.min(Math.max(count, 3), 8))
  const total = boxes.reduce((sum, box) => sum + box.pixels.length, 0)
  const colors = boxes
    .map((box) => ({ hex: averageHex(box.pixels), ratio: box.pixels.length / total }))
    .filter((item) => !isBackground(item, total))

  const preview = canvas.toDataURL('image/jpeg', 0.72)
  return { colors: sortByRatio(dedupe(colors)).slice(0, count), preview }
}

async function loadBitmap(file: File): Promise<ImageBitmap> {
  if ('createImageBitmap' in window) {
    try {
      return await createImageBitmap(file)
    } catch {
      // 某些浏览器对 HEIC 等格式会抛错，回退到 <img>
    }
  }
  const url = URL.createObjectURL(file)
  try {
    const image = await new Promise<HTMLImageElement>((resolve, reject) => {
      const element = new Image()
      element.onload = () => resolve(element)
      element.onerror = () => reject(new Error('无法读取这张照片'))
      element.src = url
    })
    return await createImageBitmap(image)
  } finally {
    URL.revokeObjectURL(url)
  }
}

/** 背景判定：接近纯白且占比很高时（常见于桌面拍摄）丢弃。 */
function isBackground(color: ExtractedColor, total: number): boolean {
  const { l, s } = hexToHsl(color.hex)
  return color.ratio > 0.35 && l > 0.9 && s < 0.12 && total > 0
}

function dedupe(colors: ExtractedColor[]): ExtractedColor[] {
  const merged: ExtractedColor[] = []
  for (const color of colors) {
    const near = merged.find((item) => hexDistance(item.hex, color.hex) < 24)
    if (near) {
      near.ratio += color.ratio
      continue
    }
    merged.push({ ...color })
  }
  return merged
}

function sortByRatio(colors: ExtractedColor[]): ExtractedColor[] {
  return [...colors].sort((left, right) => right.ratio - left.ratio)
}

type Box = { pixels: Pixel[] }

/** 中位切分：每次挑通道跨度最大的盒子沿中位数切开，直到达到目标色数。 */
function medianCut(pixels: Pixel[], target: number): Box[] {
  let boxes: Box[] = [{ pixels }]
  while (boxes.length < target) {
    let candidateIndex = -1
    let candidateRange = 0
    boxes.forEach((box, index) => {
      if (box.pixels.length < 2) return
      const range = widestRange(box.pixels).range
      if (range > candidateRange) {
        candidateRange = range
        candidateIndex = index
      }
    })
    if (candidateIndex < 0) break
    const box = boxes[candidateIndex]
    const { channel } = widestRange(box.pixels)
    const sorted = [...box.pixels].sort((left, right) => left[channel] - right[channel])
    const middle = Math.floor(sorted.length / 2)
    boxes = [...boxes.slice(0, candidateIndex), { pixels: sorted.slice(0, middle) }, { pixels: sorted.slice(middle) }, ...boxes.slice(candidateIndex + 1)]
  }
  return boxes.filter((box) => box.pixels.length > 0)
}

function widestRange(pixels: Pixel[]): { channel: 0 | 1 | 2; range: number } {
  let bestChannel: 0 | 1 | 2 = 0
  let bestRange = -1
  for (const channel of [0, 1, 2] as const) {
    let min = 255
    let max = 0
    for (const pixel of pixels) {
      if (pixel[channel] < min) min = pixel[channel]
      if (pixel[channel] > max) max = pixel[channel]
    }
    if (max - min > bestRange) {
      bestRange = max - min
      bestChannel = channel
    }
  }
  return { channel: bestChannel, range: bestRange }
}

function averageHex(pixels: Pixel[]): string {
  let red = 0
  let green = 0
  let blue = 0
  for (const pixel of pixels) {
    red += pixel[0]
    green += pixel[1]
    blue += pixel[2]
  }
  const count = pixels.length || 1
  return '#' + [red / count, green / count, blue / count].map((value) => Math.round(value).toString(16).padStart(2, '0')).join('').toUpperCase()
}

export function hexToHsl(hex: string): { h: number; s: number; l: number } {
  const red = parseInt(hex.slice(1, 3), 16) / 255
  const green = parseInt(hex.slice(3, 5), 16) / 255
  const blue = parseInt(hex.slice(5, 7), 16) / 255
  const max = Math.max(red, green, blue)
  const min = Math.min(red, green, blue)
  const delta = max - min
  let hue = 0
  if (delta) hue = max === red ? (60 * (((green - blue) / delta + 6) % 6)) : max === green ? 60 * ((blue - red) / delta + 2) : 60 * ((red - green) / delta + 4)
  const lightness = (max + min) / 2
  const saturation = delta === 0 ? 0 : delta / (1 - Math.abs(2 * lightness - 1))
  return { h: hue, s: saturation, l: lightness }
}

function hexDistance(left: string, right: string): number {
  const [lr, lg, lb] = [left.slice(1, 3), left.slice(3, 5), left.slice(5, 7)].map((part) => parseInt(part, 16))
  const [rr, rg, rb] = [right.slice(1, 3), right.slice(3, 5), right.slice(5, 7)].map((part) => parseInt(part, 16))
  return Math.sqrt((lr - rr) ** 2 + (lg - rg) ** 2 + (lb - rb) ** 2)
}
