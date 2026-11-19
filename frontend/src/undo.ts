export type UndoAction = { label: string; undo: () => Promise<void> }

let actions: UndoAction[] = []
const listeners = new Set<() => void>()
const notify = () => listeners.forEach((listener) => listener())

export function pushUndo(action: UndoAction) { actions = [...actions.slice(-19), action]; notify() }
export function latestUndo() { return actions.at(-1) }
export function discardLatestUndo() { actions = actions.slice(0, -1); notify() }
export function subscribeUndo(listener: () => void) { listeners.add(listener); return () => { listeners.delete(listener) } }
