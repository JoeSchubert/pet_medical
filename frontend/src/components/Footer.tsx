export default function Footer() {
  const year = new Date().getFullYear()
  return (
    <footer className="footer">
      <div className="footer-brand">Pet Medical</div>
      <div className="footer-copy">© {year} Pet Medical. Your pet's health portfolio.</div>
    </footer>
  )
}
