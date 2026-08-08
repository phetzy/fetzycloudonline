import {
	clampIndex,
	firstSectionOfTab,
	listStatus,
	promptSegments,
	visibleSections
} from './selectors'
import { SECTIONS } from './content'

test('with no filter, shows only the active tab', () => {
	expect(visibleSections('projects', '', false).map((s) => s.id)).toEqual([
		'mapwright',
		'3dpass',
		'transfer',
		'platform',
		'hat',
		'oss'
	])
	expect(visibleSections('work', '', false).map((s) => s.id)).toEqual(['mapwright-work', 'c1', 'stack'])
})

test('while the filter is open, searches every section regardless of tab', () => {
	const hits = visibleSections('readme', 'stack', true).map((s) => s.id)
	expect(hits).toContain('stack')
})

test('with a filter string but the filter closed, stays within the active tab', () => {
	expect(visibleSections('readme', 'stack', false).map((s) => s.id)).toEqual([])
})

test('filter matches label and title, case-insensitively', () => {
	expect(visibleSections('projects', 'MAPWRIGHT', true).map((s) => s.id)).toEqual([
		'mapwright',
		'mapwright-work'
	])
	expect(visibleSections('projects', 'secure element', true).map((s) => s.id)).toEqual(['hat'])
})

test('filter with no match returns nothing', () => {
	expect(visibleSections('projects', 'zzzzqqq', true)).toEqual([])
})

test('list status reports position and total', () => {
	const visible = visibleSections('projects', '', false)
	expect(listStatus(visible, 'mapwright', '')).toBe('1/6')
	expect(listStatus(visible, 'oss', '')).toBe('6/6')
})

test('list status marks an active filter', () => {
	const visible = visibleSections('projects', 'map', true)
	expect(listStatus(visible, 'mapwright', 'map')).toBe('1/2 filtered')
})

test('list status reports no match on an empty list', () => {
	expect(listStatus([], 'mapwright', 'zzz')).toBe('no match')
})

test('prompt always carries directory and branch', () => {
	const readme = SECTIONS.find((s) => s.id === 'readme')!
	const segs = promptSegments(readme)
	expect(segs[0].text).toBe('~/site/README.md')
	expect(segs[1].text).toContain('main')
	expect(segs).toHaveLength(2)
})

test('prompt adds a language module only for sections that define one', () => {
	const mapwright = SECTIONS.find((s) => s.id === 'mapwright')!
	expect(promptSegments(mapwright)).toHaveLength(3)
	expect(promptSegments(mapwright)[2].text).toContain('docker')

	const contact = SECTIONS.find((s) => s.id === 'contact')!
	expect(promptSegments(contact)).toHaveLength(2)
})

test('first section of a tab', () => {
	expect(firstSectionOfTab('projects').id).toBe('mapwright')
	expect(firstSectionOfTab('work').id).toBe('mapwright-work')
})

test('clampIndex stops at both ends and does not wrap', () => {
	expect(clampIndex(0, -1, 5)).toBe(0)
	expect(clampIndex(4, 1, 5)).toBe(4)
	expect(clampIndex(0, 99, 5)).toBe(4)
	expect(clampIndex(4, -99, 5)).toBe(0)
	expect(clampIndex(1, 1, 5)).toBe(2)
})
