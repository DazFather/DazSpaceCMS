# DazSpaceCMS

Customizable and easy pure Golang server + Content Management System made to serve blogs; see it on my personal site: [DazSpace.codes](https://DazSpace.codes/).
The project is already usable but still **Work In Progress**
> I would suggest to wait for the first release to use it. Everyone is free to use it for their personal or commercial use but selling this piece of software is forbidden (see LICENSE).


## Easy to use

**Write in markdown:** no need for hundreds of tags inside a nest-monster. Just write in your favorite editor it

**Drag and drop:** once you have your markdown article simply save it inside the article folder in the contents directories (contents > articles) and it will automagically rename it and create an HTML file connected to it inside the blog folders


## Ready to use

**Install:** To use it just clone this repo, position yourself inside the folder and run this command on the shell: go build.
Start the executable file just created. It will automagically create all the necessary folders and settings files

**Set up:** Inside the "setting" folder you can find a file called "site.JSON" open it up and fill it using your own datas. This operation (changing a setting files) is the only one that require restarting the server


## Really customizable

Thanks to the **Template engine** you can edit as you want all the .tmpl files that you can find inside "settings > templates" and use the placeholders where you need.

It also very easily **support frameworks and external libraries** like Sass, Tailwind or Bootstrap. Or any other that allow you to have am html template and a css on their folders insider the resources directories.

**Attach multiple extra stylesheets and JS files** even for single pages, by simply write:
[!SCRIPT](link)(link2)(link3)... for scripts
[!STYLE](link)(link2)(link3)... for stylesheets
> if they are internal then you don't need to specify the full path but just the full name.


## Live changes

Real time changes for JS, CSS and template files.

Just edit your markdown article and save it to see in the generated HTML file (and online) the same changes.
> Service like Cloudflare could interfere with this feature. Make sure to purge the cache and activate the developer mode to see the changes in your site


## Feature list

- Create any article easily in Markdown and it will transform into HTML thanks to the template system and render it on the blog page
- You can attach to specific articles their own CSS or JS sheets
- Easy support for most of CSS frameworks like Sass or Tailwind
- Live changes when you edit your markdown articles
- Automatic backup of the html file when you delete the original markdown article
- Automatic generation a SEO-friendly links for each article based on title + date
- No collision for articles with the same title
- Live changes for any resources (CSS or JS sheets, images or HTML templates)
- Ability to change settings like the paths of contents and resources directories and your site infos
