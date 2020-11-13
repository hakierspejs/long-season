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

/*!
 * Just the bare necessities of state management 
 * (c) 2018 Google LLC, Apache-2.0 License
 *
 * https://gist.github.com/developit/a0430c500f5559b715c2dddf9c40948d
 *
 * @param {Any} arg argument
 */
window.valoo = function (v, cb) {
  cb = cb || [];
  return function (c) {
    if (c === void 0) return v;
    if (c.call) return cb.splice.bind(cb, cb.push(c) - 1, 1, null);
    v = c;
    for (var i = 0, l = cb.length; i < l; i++) {
      cb[i] && cb[i](v);
    }
  };
};
