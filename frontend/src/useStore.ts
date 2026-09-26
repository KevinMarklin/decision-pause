import { useContext } from 'react'
import { StoreContext, type Store } from './storeContext'

export function useStore(): Store {
  const ctx = useContext(StoreContext)
  if (!ctx) throw new Error('useStore must be used within StoreProvider')
  return ctx
}
