import { SECTIONS, type Section, type TabId } from './content'

/** Starship language modules, per section, in Macchiato-mapped colors. */
const LANG: Record<string, { text: string; color: string }> = {
	mapwright: { text: ' docker', color: '#8aadf4' },
	transfer: { text: ' go', color: '#91d7e3' },
	platform: { text: ' python', color: '#eed49f' },
	hat: { text: ' i2c', color: '#8bd5ca' },
	c1: { text: ' go', color: '#91d7e3' },
	stack: { text: ' node', color: '#a6da95' }
}

/**
 * While the filter input is open the search covers every section, not just the
 * active tab — selecting a match is what moves you to another tab. With a
 * filter string but the input closed, the active tab still bounds the list.
 */
export function visibleSections(tab: TabId, filter: string, filtering: boolean): Section[] {
	const q = filter.trim().toLowerCase()
	const inTab = SECTIONS.filter((s) => s.tab === tab)
	if (!q) return inTab
	const pool = filtering ? SECTIONS : inTab
	return pool.filter((s) => `${s.label} ${s.title}`.toLowerCase().includes(q))
}

export function listStatus(visible: Section[], selectedId: string, filter: string): string {
	if (visible.length === 0) return 'no match'
	const i = visible.findIndex((s) => s.id === selectedId)
	return `${i + 1}/${visible.length}${filter ? ' filtered' : ''}`
}

export function promptSegments(section: Section): { text: string; color: string }[] {
	const segs = [
		{ text: section.path, color: '#b7bdf8' },
		{ text: '\u{E725} main', color: '#c6a0f6' }
	]
	const lang = LANG[section.id]
	if (lang) segs.push(lang)
	return segs
}

export function firstSectionOfTab(tab: TabId): Section {
	return SECTIONS.filter((s) => s.tab === tab)[0] ?? SECTIONS[0]
}

export function clampIndex(current: number, delta: number, length: number): number {
	return Math.max(0, Math.min(current + delta, length - 1))
}
