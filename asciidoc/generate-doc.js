const jsdoc2md = require('jsdoc-to-markdown');
const fs = require('fs');

const title = process.argv[2];
const file = process.argv[3];

// 👇 read template file manually
const template = fs.readFileSync('asciidoc/api-template.hbs', 'utf8');

(async () => {
    const content = await jsdoc2md.render({
        files: file,
        template: template
    });

    const header = `= ${title}
:toc:
:toclevels: 2

`;

    fs.writeFileSync(process.stdout.fd, header + content);
})();
