import { PrismLight } from 'react-syntax-highlighter'
import python from 'react-syntax-highlighter/dist/esm/languages/prism/python'
import typescript from 'react-syntax-highlighter/dist/esm/languages/prism/typescript'
import javascript from 'react-syntax-highlighter/dist/esm/languages/prism/javascript'
import go from 'react-syntax-highlighter/dist/esm/languages/prism/go'
import bash from 'react-syntax-highlighter/dist/esm/languages/prism/bash'
import json from 'react-syntax-highlighter/dist/esm/languages/prism/json'

// Register only the languages the site actually renders, so the light
// build ships a fraction of the full grammar set.
PrismLight.registerLanguage('python', python)
PrismLight.registerLanguage('typescript', typescript)
PrismLight.registerLanguage('javascript', javascript)
PrismLight.registerLanguage('go', go)
PrismLight.registerLanguage('bash', bash)
PrismLight.registerLanguage('json', json)

export default PrismLight
