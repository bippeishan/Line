/**
 * 配色可读性分析：WCAG 对比度、CIE76 ΔE、红绿色盲模拟，以及真实线材照片纹理。
 * 纯前端实现，用于在配色页实时提示「这两个颜色容易数错行」。
 */

export type PaletteIssue = {
  kind: 'separation' | 'colorblind' | 'lightness'
  severity: 'warn' | 'info'
  colors: string[]
  message: string
}

export type PaletteAnalysis = {
  issues: PaletteIssue[]
  textColors: string[]
  contrastRatio: number
}

export function hexToRgb(hex: string): [number, number, number] {
  return [1, 3, 5].map((index) => parseInt(hex.slice(index, index + 2), 16)) as [number, number, number]
}

export function hexToRgba(hex: string, alpha: number): string {
  const [red, green, blue] = hexToRgb(hex)
  return `rgba(${red}, ${green}, ${blue}, ${alpha})`
}

function channelToLinear(value: number): number {
  const normalized = value / 255
  return normalized <= 0.04045 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4
}

function linearToChannel(value: number): number {
  const clamped = Math.min(1, Math.max(0, value))
  const normalized = clamped <= 0.0031308 ? clamped * 12.92 : 1.055 * clamped ** (1 / 2.4) - 0.055
  return Math.round(normalized * 255)
}

/** WCAG 相对亮度 */
export function relativeLuminance(hex: string): number {
  const [red, green, blue] = hexToRgb(hex).map(channelToLinear)
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue
}

/** WCAG 对比度，1–21 */
export function contrastRatio(left: string, right: string): number {
  const a = relativeLuminance(left)
  const b = relativeLuminance(right)
  const lighter = Math.max(a, b)
  const darker = Math.min(a, b)
  return (lighter + 0.05) / (darker + 0.05)
}

/** 在色块上放黑字还是白字 */
export function readableTextColor(hex: string): string {
  return contrastRatio(hex, '#FFFFFF') >= contrastRatio(hex, '#111111') ? '#FFFFFF' : '#111111'
}

function rgbToLab(hex: string): [number, number, number] {
  const [red, green, blue] = hexToRgb(hex).map(channelToLinear)
  const x = (0.4124 * red + 0.3576 * green + 0.1805 * blue) / 0.95047
  const y = 0.2126 * red + 0.7152 * green + 0.0722 * blue
  const z = (0.0193 * red + 0.1192 * green + 0.9505 * blue) / 1.08883
  const pivot = (value: number) => (value > 0.008856 ? Math.cbrt(value) : 7.787 * value + 16 / 116)
  const [fx, fy, fz] = [pivot(x), pivot(y), pivot(z)]
  return [116 * fy - 16, 500 * (fx - fy), 200 * (fy - fz)]
}

/** CIE76 色差，经验上 < 2.3 肉眼几乎相同，> 10 才明显能分开 */
export function deltaE(left: string, right: string): number {
  const [l1, a1, b1] = rgbToLab(left)
  const [l2, a2, b2] = rgbToLab(right)
  return Math.sqrt((l1 - l2) ** 2 + (a1 - a2) ** 2 + (b1 - b2) ** 2)
}

const blindnessMatrices = {
  protanopia: [0.56667, 0.43333, 0, 0.55833, 0.44167, 0, 0, 0.24167, 0.75833],
  deuteranopia: [0.625, 0.375, 0, 0.7, 0.3, 0, 0, 0.3, 0.7],
} as const

/** 用线性 RGB 矩阵模拟红绿色弱看到的颜色 */
export function simulateColorBlindness(hex: string, type: keyof typeof blindnessMatrices): string {
  const [red, green, blue] = hexToRgb(hex).map(channelToLinear)
  const matrix = blindnessMatrices[type]
  return (
    '#' +
    [
      matrix[0] * red + matrix[1] * green + matrix[2] * blue,
      matrix[3] * red + matrix[4] * green + matrix[5] * blue,
      matrix[6] * red + matrix[7] * green + matrix[8] * blue,
    ]
      .map((value) => linearToChannel(value).toString(16).padStart(2, '0'))
      .join('')
      .toUpperCase()
  )
}

const separationDeltaE = 14
const colorBlindDeltaE = 10
const lightnessSpreadMin = 18

/**
 * 分析一组颜色：哪些两两之间区分度不够、色弱视角下是否更难分、明暗层次是否拉开。
 * 阈值参考 CIE76 ΔE 与 WCAG 对比度经验值。
 */
export function analyzePalette(colors: string[]): PaletteAnalysis {
  const normalized = colors.map((color) => color.toUpperCase())
  const issues: PaletteIssue[] = []
  const seen = new Set<string>()

  for (let left = 0; left < normalized.length; left++) {
    for (let right = left + 1; right < normalized.length; right++) {
      const pair = [normalized[left], normalized[right]]
      const key = pair.join('|')
      if (seen.has(key)) continue
      seen.add(key)

      const normalGap = deltaE(pair[0], pair[1])
      if (normalGap < separationDeltaE) {
        issues.push({
          kind: 'separation',
          severity: 'warn',
          colors: pair,
          message: `${pair[0]} 与 ${pair[1]} 色差只有 ${normalGap.toFixed(0)}，钩起来容易看混，建议拉开明暗或换一个色。`,
        })
        continue
      }
      for (const type of ['deuteranopia', 'protanopia'] as const) {
        const gap = deltaE(simulateColorBlindness(pair[0], type), simulateColorBlindness(pair[1], type))
        if (gap < colorBlindDeltaE) {
          issues.push({
            kind: 'colorblind',
            severity: 'warn',
            colors: pair,
            message: `在${type === 'deuteranopia' ? '绿色弱' : '红色弱'}视角下，${pair[0]} 与 ${pair[1]} 几乎一样（色差 ${gap.toFixed(0)}）。`,
          })
          break
        }
      }
    }
  }

  const lightness = normalized.map((color) => rgbToLab(color)[0])
  const spread = Math.max(...lightness) - Math.min(...lightness)
  if (normalized.length >= 2 && spread < lightnessSpreadMin) {
    issues.push({
      kind: 'lightness',
      severity: 'info',
      colors: normalized,
      message: `这组颜色的明暗跨度只有 ${spread.toFixed(0)}，行与行之间不太容易一眼分清，可以加一个更深或更浅的颜色。`,
    })
  }

  const ratios = normalized.map((color) => contrastRatio(color, '#FFFFFF'))
  return {
    issues,
    textColors: normalized.map((color) => readableTextColor(color)),
    contrastRatio: Math.max(...ratios),
  }
}

const textureCache = new Map<string, string>()

/** 把线材照片裁成中心方块当色卡纹理，结果按 URL 缓存。 */
export async function yarnTexture(source: string): Promise<string | null> {
  if (!source) return null
  const cached = textureCache.get(source)
  if (cached !== undefined) return cached || null
  try {
    const image = await new Promise<HTMLImageElement>((resolve, reject) => {
      const element = new Image()
      element.crossOrigin = 'same-origin'
      element.onload = () => resolve(element)
      element.onerror = () => reject(new Error('无法读取线材照片'))
      element.src = source
    })
    const size = 96
    const side = Math.min(image.width, image.height)
    const canvas = document.createElement('canvas')
    canvas.width = size
    canvas.height = size
    const context = canvas.getContext('2d')
    if (!context) throw new Error('无法绘制线材照片')
    context.drawImage(image, (image.width - side) / 2, (image.height - side) / 2, side, side, 0, 0, size, size)
    const dataUrl = canvas.toDataURL('image/jpeg', 0.78)
    textureCache.set(source, dataUrl)
    return dataUrl
  } catch {
    textureCache.set(source, '')
    return null
  }
}
