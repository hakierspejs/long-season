/*!
 * Run event after the DOM is ready
 * (c) 2017 Chris Ferdinandi, MIT License, https://gomakethings.com
 * @param  {Function} fn Callback function
 */
window.ready = function (fn) {
  // Sanity check
  if (typeof fn !== "function") return;

  // If document is already loaded, run method
  if (
    document.readyState === "interactive" ||
    document.readyState === "complete"
  ) {
    return fn();
  }

  // Otherwise, wait until document is loaded
  document.addEventListener("DOMContentLoaded", fn, false);
};

/*!
 * The identity function takes one argument and
 * returns that argument.
 * @param {Any} arg argument
 */
window.id = (arg) => arg;
