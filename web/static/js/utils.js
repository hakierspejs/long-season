/*!
 * The identity function takes one argument and
 * returns that argument.
 * @param {Any} arg argument
 */
const id = (arg) => arg;

/*!
 * Just the bare necessities of state management
 * (c) 2018 Google LLC, Apache-2.0 License
 *
 * https://gist.github.com/developit/a0430c500f5559b715c2dddf9c40948d
 *
 * @param {Any} v initial value
 * @param {Array.<Function>} cb array with callback functions
 */
function valoo(v, cb) {
  cb = cb || [];
  return function (c) {
    if (c === void 0) return v;
    if (c.call) return cb.splice.bind(cb, cb.push(c) - 1, 1, null);
    v = c;
    for (var i = 0, l = cb.length; i < l; i++) {
      cb[i] && cb[i](v);
    }
  };
}

/*!
 * Creates new HTMLElement with given tag, props and children.
 *
 * Implemented to use instead of HTML templates, because you can
 * use regulra JavaScript to generate nodes.
 *
 * @param {String} tag Html tag like: div, p, b etc
 * @param {Object} props Dictionary with properties for new element
 * @param {Array.<HTMLElement|String>} children
 * @return {HTMLElement} Created node
 */

function el(tag, props, ...children) {
  if (typeof tag === "undefined") return false;

  // Pass empty string if children is undefined.
  if (typeof children === "undefined") children = [""];

  const result = document.createElement(tag);

  if (typeof props === "object") {
    for (const key in props) {
      let eventName = key.match(/^on([A-Z]\w+)$/);
      // If key matches some event name, add event listener with given function
      // for matched event.
      if (eventName) {
        result.addEventListener(eventName[1].toLowerCase(), props[key]);
      } // Otherwise set regular attribute.
      else {
        result.setAttribute(key, props[key]);
      }
    }
  }

  // For each child, add to new element. If child is not
  // HTMLElement, create new text node.
  children.forEach((child) => {
    result.appendChild(
      child instanceof HTMLElement ? child : document.createTextNode(child),
    );
  });

  return result;
}

/*
 * withErr awaits for given promises and returns result
 * and null if there were error or null and error if some
 * error had occurred.
 */
async function withErr(promise) {
  try {
    let res = await promise;
    return [res, null];
  } catch (error) {
    return [null, error];
  }
}

/*
 * empty removes all child nodes from given node.
 */
function empty(node) {
  while (node.firstChild) {
    node.removeChild(node.firstChild);
  }
}

/*
 * render replaces all children nodes of given parent
 * with given target node or array of nodes.
 */
function render(parentNode, target) {
  empty(parentNode);

  if (Array.isArray(target)) {
    target.forEach((node, _) => {
      parentNode.append(node);
    });
  } else {
    parentNode.append(target);
  }
}

/*
 * main is asynchronous wrapper for given function.
 * It can be used in scripts for writing convenient main
 * functions.
 */
async function main(wrapped) {
  wrapped();
}

export { el, empty, id, main, render, valoo, withErr };
