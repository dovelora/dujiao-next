import DOMPurify from 'dompurify'

const blockTags = new Set([
  'p', 'div', 'section', 'article', 'header', 'footer', 'aside', 'blockquote', 'pre',
  'ul', 'ol', 'li', 'dl', 'dt', 'dd', 'table', 'thead', 'tbody', 'tfoot', 'tr', 'td', 'th',
  'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
])
const textTags = [...blockTags, 'br', 'hr', 'span', 'a', 'b', 'strong', 'i', 'em', 'u', 's', 'del', 'code', 'small', 'sub', 'sup']
// Detection only: parsing and entity decoding are delegated to DOMPurify's DOM.
const htmlTag = new RegExp(`</?(?:${[...textTags, 'script', 'style', 'img', 'svg', 'math', 'iframe', 'object', 'embed', 'template', 'noscript', 'form', 'input', 'textarea', 'button', 'select', 'link', 'meta', 'base', 'title', 'head', 'body', 'html'].join('|')})(?=[\\s/>])`, 'i')

/** Normalize imported descriptions for text interpolation; never use the result with v-html. */
export function getProductDescriptionText(description: string): string {
  if (typeof description !== 'string' || !description) return ''
  // Unknown placeholders such as <token> are ordinary text, not HTML elements.
  const source = htmlTag.test(description)
    ? description
    : description.replace(/</g, '&lt;').replace(/>/g, '&gt;')
  const fragment = DOMPurify.sanitize(source, {
    ALLOWED_TAGS: textTags,
    ALLOWED_ATTR: [],
    ADD_FORBID_CONTENTS: ['object', 'embed'],
    RETURN_DOM_FRAGMENT: true,
  })

  const readText = (node: Node): string => {
    if (node.nodeType === Node.TEXT_NODE) return node.textContent || ''
    const tag = node.nodeType === Node.ELEMENT_NODE ? (node as Element).tagName.toLowerCase() : ''
    if (tag === 'br' || tag === 'hr') return '\n'
    const text = Array.from(node.childNodes, readText).join('')
    return blockTags.has(tag) ? `\n${text}\n` : text
  }

  return readText(fragment)
    .replace(/\r\n?/g, '\n')
    .replace(/\u00a0/g, ' ')
    .replace(/[\t ]*\n[\t ]*/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}
