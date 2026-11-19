import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles.css'
import App from './App'
import { installOfflineSupport } from './offline'

installOfflineSupport()
createRoot(document.getElementById('root')!).render(<StrictMode><App /></StrictMode>)
