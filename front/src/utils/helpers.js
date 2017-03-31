export const getKeyByValue = (object, value) => {
    return Object.keys(object).find(key => object[key] === value);
}

export const toFirstLowerCase = (string) => {
    return string.charAt(0).toLowerCase() + string.slice(1);
}