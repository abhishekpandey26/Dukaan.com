// Shows an error or success banner. Renders nothing when text is empty.
export default function Message({ type = 'error', text }) {
  if (!text) return null
  return <div className={`message message-${type}`}>{text}</div>
}
