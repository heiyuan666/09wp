import '@semi-ui-styles/semi.min.css'

import SearchPage from './pages/SearchPage'
import type { SearchBridge } from './useSearchPage'

export default function SearchReact(bridge: SearchBridge) {
  return <SearchPage {...bridge} />
}
