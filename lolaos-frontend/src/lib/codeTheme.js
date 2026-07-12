/**
 * A strictly monochrome syntax-highlighting theme for react-syntax-highlighter.
 * Per branding.md there is no hue anywhere — emphasis in code comes only
 * from lightness contrast across the grayscale ramp. Rendered on a
 * gray-950 terminal surface in both light and dark themes.
 */
const base = {
  color: '#e5e5e5',
  background: 'transparent',
  fontFamily: "'JetBrains Mono', ui-monospace, monospace",
  fontSize: '0.85rem',
  lineHeight: '1.7',
  direction: 'ltr',
  textAlign: 'left',
  whiteSpace: 'pre',
  wordSpacing: 'normal',
  wordBreak: 'normal',
  tabSize: 2,
}

export const monoCodeTheme = {
  'code[class*="language-"]': base,
  'pre[class*="language-"]': { ...base, margin: 0, padding: 0, overflow: 'auto' },
  comment: { color: '#737373', fontStyle: 'italic' },
  prolog: { color: '#737373' },
  doctype: { color: '#737373' },
  cdata: { color: '#737373' },
  punctuation: { color: '#a3a3a3' },
  namespace: { opacity: 0.7 },
  property: { color: '#d4d4d4' },
  tag: { color: '#f5f5f5' },
  boolean: { color: '#fafafa' },
  number: { color: '#d4d4d4' },
  constant: { color: '#fafafa' },
  symbol: { color: '#d4d4d4' },
  deleted: { color: '#737373' },
  selector: { color: '#d4d4d4' },
  'attr-name': { color: '#d4d4d4' },
  string: { color: '#a3a3a3' },
  char: { color: '#a3a3a3' },
  builtin: { color: '#f5f5f5' },
  inserted: { color: '#d4d4d4' },
  operator: { color: '#a3a3a3' },
  entity: { color: '#d4d4d4', cursor: 'help' },
  url: { color: '#a3a3a3' },
  '.language-css .token.string': { color: '#a3a3a3' },
  '.style .token.string': { color: '#a3a3a3' },
  atrule: { color: '#f5f5f5' },
  'attr-value': { color: '#a3a3a3' },
  keyword: { color: '#fafafa', fontWeight: '600' },
  function: { color: '#f5f5f5' },
  'class-name': { color: '#fafafa' },
  regex: { color: '#a3a3a3' },
  important: { color: '#fafafa', fontWeight: '700' },
  variable: { color: '#d4d4d4' },
  bold: { fontWeight: '700' },
  italic: { fontStyle: 'italic' },
}

export default monoCodeTheme
