// paste in browser console at https://paimon.moe/characters/

let chars = []
let elems = document.getElementsByClassName("p-1")
for (let i = 0; i < elems.length; i++) {
    chars.push(elems[i].textContent)
}
console.log('{\n"' + chars.join('",\n"') + '",\n}')

// Also, replace all Traveler(x) with just Traveler