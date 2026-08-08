import { SECTIONS, TABS, type TabId } from './content'

test('eleven sections in the documented order', () => {
	expect(SECTIONS.map((s) => s.id)).toEqual([
		'readme',
		'mapwright',
		'3dpass',
		'transfer',
		'platform',
		'hat',
		'oss',
		'mapwright-work',
		'c1',
		'stack',
		'contact'
	])
})

test('four tabs', () => {
	expect(TABS.map((t) => t.id)).toEqual(['readme', 'projects', 'work', 'contact'])
})

test('section ids are unique', () => {
	expect(new Set(SECTIONS.map((s) => s.id)).size).toBe(SECTIONS.length)
})

test('every section belongs to a known tab', () => {
	const known = new Set<TabId>(TABS.map((t) => t.id))
	for (const s of SECTIONS) expect(known.has(s.tab)).toBe(true)
})

test('every section has the required fields', () => {
	for (const s of SECTIONS) {
		expect(s.label).not.toBe('')
		expect(s.path).not.toBe('')
		expect(s.title).not.toBe('')
		expect(Array.isArray(s.paras)).toBe(true)
		expect(Array.isArray(s.rows)).toBe(true)
		expect(Array.isArray(s.links)).toBe(true)
	}
})

test('every link has a label and a non-empty href', () => {
	for (const s of SECTIONS) {
		for (const l of s.links) {
			expect(l.label).not.toBe('')
			expect(l.href).not.toBe('')
		}
	}
})

test('the contact section carries three links', () => {
	const contact = SECTIONS.find((s) => s.id === 'contact')!
	expect(contact.links.map((l) => l.href)).toEqual([
		'mailto:david.j.fetzer@gmail.com',
		'https://github.com/phetzy',
		'https://linkedin.com/in/fetzy'
	])
})
