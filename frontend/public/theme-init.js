// Applies the saved theme (light/dark/system) to <html> before the first
// paint, so an explicit override never flashes the wrong theme for a frame
// while React/lib/theme.ts loads. "system" needs no attribute at all —
// index.css's prefers-color-scheme media query already handles that case
// with zero JS.
(function () {
  var stored = localStorage.getItem("financist:theme");
  if (stored === "light" || stored === "dark") {
    document.documentElement.setAttribute("data-theme", stored);
  }
  var isDark =
    stored === "dark" || (stored !== "light" && window.matchMedia("(prefers-color-scheme: dark)").matches);
  var meta = document.querySelector('meta[name="theme-color"]');
  // Kept in sync by hand with index.css's --surface-0 light/dark values.
  if (meta) meta.setAttribute("content", isDark ? "#0d0d0d" : "#f9f9f7");
})();
