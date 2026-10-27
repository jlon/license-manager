import { describe, expect, it } from 'vitest'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, resolve } from 'node:path'

/**
 * i18n 覆盖率回归。
 *
 * 背景：设备管理页的路由、视图、侧边栏图标、面包屑都先落地了，唯独漏了两类文案
 * （`navigation.menu.devices` 与整块 `pages.devices`）。结果是菜单标签渲染成空串、
 * 看起来像"设备管理入口消失了"，进入页面后所有列头与提示也全是空白 key。
 *
 * 这类缺陷不会让构建失败，只会在界面上静默退化，所以必须由测试守住：
 * 1. 视图里静态引用的 `t('...')` 必须在所有语言里都能解析；
 * 2. 动态拼接的 key（如 `pages.devices.list.status.${status}`）单独列出；
 * 3. 各语言的 key 集合必须一致，避免只有中文可用。
 */

const SRC = resolve(__dirname, '../')
const LOCALES = ['zh', 'en']

function loadLocale(locale: string): Record<string, unknown> {
  return JSON.parse(readFileSync(join(SRC, 'i18n/locales', `${locale}.json`), 'utf8'))
}

/** 按点号路径取值；任一层缺失返回 undefined。 */
function valueAt(tree: Record<string, unknown>, dotted: string): unknown {
  return dotted.split('.').reduce<unknown>((node, part) => {
    if (node === null || typeof node !== 'object') return undefined
    return (node as Record<string, unknown>)[part]
  }, tree)
}

/** 收集目录下所有 .vue/.ts 中静态引用的 i18n key。 */
function collectReferencedKeys(dir: string, found = new Set<string>()): Set<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry)
    if (statSync(full).isDirectory()) {
      collectReferencedKeys(full, found)
      continue
    }
    if (!/\.(vue|ts)$/.test(entry)) continue
    const text = readFileSync(full, 'utf8')
    for (const match of text.matchAll(/\bt\(\s*'([^']+)'/g)) {
      found.add(match[1])
    }
  }
  return found
}

/** 动态 key：设备状态由数据决定，静态扫描抓不到，这里显式声明。 */
const DYNAMIC_KEYS = ['active', 'inactive', 'revoked'].flatMap(status => [
  `pages.devices.list.status.${status}`,
])

/** 本轮修复涉及的设备管理文案，缺任何一条都会让界面出现空白。 */
const DEVICE_KEYS = [
  'navigation.menu.devices',
  'navigation.breadcrumb.devices',
  'pages.devices.list.title',
  'pages.devices.list.description',
  'pages.devices.list.filter.statusPlaceholder',
  'pages.devices.list.filter.onlinePlaceholder',
  'pages.devices.list.filter.query',
  'pages.devices.list.filter.reset',
  'pages.devices.list.online.online',
  'pages.devices.list.online.offline',
  'pages.devices.list.table.licenseKey',
  'pages.devices.list.table.fingerprint',
  'pages.devices.list.table.customer',
  'pages.devices.list.table.authorizationCode',
  'pages.devices.list.table.status',
  'pages.devices.list.table.online',
  'pages.devices.list.table.activatedAt',
  'pages.devices.list.table.lastHeartbeat',
  'pages.devices.list.table.lastOnlineIp',
  'pages.devices.list.table.action',
  'pages.devices.list.actions.refresh',
  'pages.devices.list.actions.retry',
  'pages.devices.list.actions.copy',
  'pages.devices.list.actions.viewDetail',
  'pages.devices.list.message.loadError',
  'pages.devices.list.message.copySuccess',
  'pages.devices.list.message.copyFailure',
]

describe('i18n 设备管理文案', () => {
  it.each(LOCALES)('%s 覆盖全部设备管理 key', locale => {
    const tree = loadLocale(locale)
    const missing = [...DEVICE_KEYS, ...DYNAMIC_KEYS].filter(
      key => typeof valueAt(tree, key) !== 'string' || String(valueAt(tree, key)).trim() === '',
    )
    expect(missing).toEqual([])
  })

  it('侧边栏文案非空：空串会让菜单项看起来消失了', () => {
    for (const locale of LOCALES) {
      const label = valueAt(loadLocale(locale), 'navigation.menu.devices')
      expect(typeof label).toBe('string')
      expect((label as string).trim().length).toBeGreaterThan(0)
    }
  })

  it('视图静态引用的所有 key 在所有语言里都能解析', () => {
    const referenced = [...collectReferencedKeys(SRC)]
    // 只校验设备相关与通用导航，避免历史遗留页面的既有缺口阻断本次回归。
    const scope = referenced.filter(
      key => key.includes('.devices') || key.startsWith('navigation.menu.'),
    )
    expect(scope.length).toBeGreaterThan(0)

    for (const locale of LOCALES) {
      const tree = loadLocale(locale)
      const missing = scope.filter(key => valueAt(tree, key) === undefined)
      expect(missing, `${locale} 缺失: ${missing.join(', ')}`).toEqual([])
    }
  })

  it('各语言 key 集合一致，避免只有单一语言可用', () => {
    const shape = (tree: unknown, prefix = ''): string[] => {
      if (tree === null || typeof tree !== 'object') return [prefix]
      return Object.entries(tree as Record<string, unknown>).flatMap(([key, value]) =>
        shape(value, prefix ? `${prefix}.${key}` : key),
      )
    }
    const [first, ...rest] = LOCALES.map(locale => shape(loadLocale(locale)))
    for (const other of rest) {
      expect(other.filter(key => !first.includes(key))).toEqual([])
      expect(first.filter(key => !other.includes(key))).toEqual([])
    }
  })
})
