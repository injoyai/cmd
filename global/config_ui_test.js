// No browser dependency: exercise form state and serialization with a small DOM fixture.
const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

function setup() {
    let document;
    class Element {
        constructor(tag) {
            this.tag = tag;
            this.children = [];
            this.attributes = {};
            this.listeners = {};
            this.dataset = {};
            this.style = {};
            this.value = '';
            this.className = '';
            this.classList = { add: name => { this.className += ' ' + name; } };
        }
        append(...children) { children.forEach(child => this.appendChild(child)); }
        appendChild(child) { child.parentElement = this; this.children.push(child); return child; }
        replaceChildren(...children) { this.children = []; this.append(...children); }
        remove() { this.parentElement.children.splice(this.parentElement.children.indexOf(this), 1); }
        get nextElementSibling() { return this.parentElement.children[this.parentElement.children.indexOf(this) + 1]; }
        get previousElementSibling() { return this.parentElement.children[this.parentElement.children.indexOf(this) - 1]; }
        setAttribute(key, value) { this.attributes[key] = value; }
        getAttribute(key) { return this.attributes[key] ?? null; }
        removeAttribute(key) { delete this.attributes[key]; }
        addEventListener(type, callback) { (this.listeners[type] ||= []).push(callback); }
        dispatchEvent(event) {
            (this.listeners[event.type] || []).forEach(callback => callback.call(this, event));
            if (event.bubbles && this.parentElement) this.parentElement.dispatchEvent(event);
        }
        querySelectorAll(selector) {
            const result = [];
            for (const child of this.children) {
                if (selector.split(',').some(s => s[0] === '.' ? child.className.split(' ').includes(s.slice(1)) : child.tag === s)) result.push(child);
                result.push(...child.querySelectorAll(selector));
            }
            return result;
        }
        querySelector(selector) { return this.querySelectorAll(selector)[0] || null; }
        focus() { document.activeElement = this; }
    }
    const root = new Element('body');
    const nodes = {};
    for (const [id, tag] of Object.entries({configForm:'form',configFields:'fieldset',saveButton:'button',saveStatus:'p'})) {
        nodes[id] = new Element(tag);
        nodes[id].id = id;
    }
    nodes.configForm.appendChild(nodes.configFields);
    root.append(nodes.configForm, nodes.saveButton, nodes.saveStatus);
    document = {
        createElement: tag => new Element(tag),
        createElementNS: (namespace, tag) => Object.assign(new Element(tag), {namespaceURI:namespace}),
        getElementById: id => root.querySelectorAll('input,textarea,button,label,section,h2,fieldset,form,p').find(node => node.id === id),
        activeElement: null
    };
    const window = new Element('window');
    const context = vm.createContext({document, window});
    const html = fs.readFileSync(path.join(__dirname, 'global.html'), 'utf8');
    vm.runInContext(html.match(/<script>([\s\S]*?)<\/script>/)[1], context);
    const fields = [
        {key:'download.dir', name:'下载目录', value:'./downloads'},
        {key:'download.noticeEnable', name:'完成通知', type:'bool', value:false},
        {key:'download.noticeText', name:'通知内容', value:''},
        {key:'proxy.addr', name:'代理地址', value:''},
        {key:'proxy.ignore', name:'忽略规则', value:''},
        {key:'resource.urls', name:'资源地址', type:'list', value:['https://one.example/{name}?a=1,b=2']},
        {key:'upload.minio.endpoint', name:'服务地址', value:'localhost:9000'},
        {key:'upload.minio.access', name:'Access Key', type:'secret', value:'example-access'},
        {key:'upload.minio.secret', name:'Secret Key', type:'secret', value:'example-secret'},
        {key:'upload.minio.bucket', name:'存储桶', value:'downloads'},
        {key:'upload.minio.rename', name:'随机文件名', type:'bool', value:false},
        {key:'open.custom', name:'快捷打开', type:'object', value:[{key:'docs', value:'/tmp/docs'}]}
    ];
    context.loadConfig(fields);
    return {context, document, window, nodes, field: id => document.getElementById(id)};
}
const plain = value => JSON.parse(JSON.stringify(value));
const click = element => element.dispatchEvent({type:'click'});
const edit = element => element.dispatchEvent({type:'input', bubbles:true});

test('five groups, native checkboxes, labels and masked credentials', () => {
    const ui = setup();
    assert.equal(ui.nodes.configFields.querySelectorAll('section').length, 5);
    assert.equal(ui.field('download.noticeEnable').type, 'checkbox');
    assert.equal(ui.nodes.saveButton.disabled, true);
    const secret = ui.field('upload.minio.secret');
    assert.equal(secret.type, 'password');
    assert.equal(ui.field('label-upload.minio.secret').htmlFor, secret.id);
    const reveal = secret.parentElement.querySelector('button');
    assert.equal(reveal.getAttribute('aria-label'), '显示 Secret Key');
    assert.equal(reveal.getAttribute('title'), '显示 Secret Key');
    assert.equal(reveal.querySelector('svg').getAttribute('aria-hidden'), 'true');
    const hiddenPath = reveal.querySelector('path').getAttribute('d');
    click(reveal);
    assert.equal(secret.type, 'text');
    assert.equal(reveal.getAttribute('aria-pressed'), 'true');
    assert.equal(reveal.getAttribute('title'), '隐藏 Secret Key');
    assert.notEqual(reveal.querySelector('path').getAttribute('d'), hiddenPath);
    click(reveal);
    assert.equal(secret.type, 'password');
    assert.equal(reveal.querySelector('path').getAttribute('d'), hiddenPath);
    for (const [selector, label] of [['.add-shortcut', '添加快捷方式'], ['.text-button', '移除此快捷方式']]) {
        const button = ui.nodes.configFields.querySelector(selector);
        assert.ok(button.querySelector('svg'));
        assert.equal(button.getAttribute('aria-label'), label);
        assert.equal(button.getAttribute('title'), label);
    }
    assert.equal(ui.nodes.saveButton.disabled, true);
});

test('lists retain commas/equals/order, explicit false and empty strings, literal map keys', async () => {
    const ui = setup();
    let saved;
    ui.window.saveToFile = async value => { saved = plain(value); };
    const list = ui.field('resource.urls');
    list.value = '\r\nhttps://new.example/{name}?x=a=b,c\r\n\r\nhttps://last.example/{name}\n';
    edit(list);
    const object = ui.nodes.configFields.querySelector('.add-shortcut').parentElement.parentElement;
    ui.context.addShortcut(object, '__proto__', '/tmp/literal', false);
    await ui.context.saveConfig();
    assert.deepEqual(saved['resource.urls'], ['https://new.example/{name}?x=a=b,c','https://last.example/{name}']);
    assert.equal(saved['download.noticeEnable'], false);
    assert.equal(saved['download.noticeText'], '');
    assert.equal(saved['upload.minio.rename'], false);
    assert.equal(saved['upload.minio.secret'], 'example-secret');
    assert.equal(saved['open.custom'].__proto__, '/tmp/literal');
    assert.equal(Object.hasOwn(saved, 'upload.minio'), false);
    assert.equal(ui.nodes.saveStatus.textContent, '已保存');
    assert.equal(ui.nodes.saveButton.disabled, true);
    list.value = '';
    edit(list);
    await ui.context.saveConfig();
    assert.deepEqual(saved['resource.urls'], []);
});

test('save failure preserves edits, reports plain text, and allows retry', async () => {
    const ui = setup();
    ui.window.saveToFile = async () => { throw new Error('disk "full" <details>'); };
    const dir = ui.field('download.dir');
    dir.value = './changed';
    edit(dir);
    await ui.context.saveConfig();
    assert.equal(dir.value, './changed');
    assert.equal(ui.nodes.saveStatus.textContent, '保存失败：disk "full" <details>');
    assert.equal(ui.nodes.saveStatus.dataset.tone, 'error');
    assert.equal(ui.nodes.configFields.disabled, false);
    assert.equal(ui.nodes.saveButton.disabled, false);
    ui.window.saveToFile = async () => {};
    await ui.context.saveConfig();
    assert.equal(ui.nodes.saveStatus.textContent, '已保存');
});

test('duplicate submit is blocked and editing stays locked until save completes', async () => {
    const ui = setup();
    let resolve;
    let calls = 0;
    ui.window.saveToFile = () => { calls++; return new Promise(done => { resolve = done; }); };
    edit(ui.field('download.dir'));
    const pending = ui.context.saveConfig();
    await ui.context.saveConfig();
    assert.equal(calls, 1);
    assert.equal(ui.nodes.configFields.disabled, true);
    assert.equal(ui.nodes.saveButton.textContent, '保存中…');
    assert.equal(ui.nodes.configForm.getAttribute('aria-busy'), 'true');
    resolve();
    await pending;
    assert.equal(ui.nodes.configFields.disabled, false);
    assert.equal(ui.nodes.configForm.getAttribute('aria-busy'), null);
});

test('duplicate/empty shortcut names block save and focus the invalid field', async () => {
    const ui = setup();
    let calls = 0;
    ui.window.saveToFile = async () => { calls++; };
    const object = ui.nodes.configFields.querySelector('.add-shortcut').parentElement.parentElement;
    ui.context.addShortcut(object, 'docs', '/tmp/duplicate', false);
    ui.context.markDirty();
    await ui.context.saveConfig();
    const name = object.querySelectorAll('.shortcut-name')[1];
    assert.equal(calls, 0);
    assert.equal(ui.document.activeElement, name);
    assert.equal(name.getAttribute('aria-invalid'), 'true');
    name.value = '';
    edit(name);
    await ui.context.saveConfig();
    assert.equal(calls, 0);
    assert.equal(name.getAttribute('aria-invalid'), 'true');
    name.value = 'other';
    edit(name);
    await ui.context.saveConfig();
    assert.equal(calls, 1);
});

test('removing the last shortcut shows empty state and preserves keyboard focus', () => {
    const ui = setup();
    const object = ui.nodes.configFields.querySelector('.add-shortcut').parentElement.parentElement;
    click(object.querySelector('.text-button'));
    assert.equal(object.querySelectorAll('.shortcut-row').length, 0);
    assert.equal(object.querySelector('.empty').hidden, false);
    assert.equal(ui.document.activeElement, object.querySelector('.add-shortcut'));
    let prevented = false;
    const closing = {type:'beforeunload', preventDefault:() => { prevented = true; }};
    ui.window.dispatchEvent(closing);
    assert.equal(prevented, false);
    assert.equal(closing.returnValue, undefined);
});
