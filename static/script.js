// JS File
'use strict';

(function () {
    const root = document.getElementById('root')
    const nav = root.appendChild(document.createElement(`nav`))
    const main = root.appendChild(document.createElement(`div`))

    nav.textContent = "navigation"
    nav.id = "nav"
    main.textContent = "main viewer"
    main.id = "content"
})()
