document.addEventListener("DOMContentLoaded", () => {
  const body = document.body;

  const savedTheme = localStorage.getItem("theme");
  if (savedTheme === "dark") {
    body.classList.add("dark-mode");
  }

  window.toggleTheme = function () {
    body.classList.toggle("dark-mode");
    const newTheme = body.classList.contains("dark-mode") ? "dark" : "light";
    localStorage.setItem("theme", newTheme);
  };

  const menuToggle = document.getElementById("menuToggle");
  if (menuToggle) {
    menuToggle.addEventListener("click", () => {
      const menu = document.getElementById("menu");
      if (menu) {
        menu.classList.toggle("hidden");
      }
    });
  }
});
