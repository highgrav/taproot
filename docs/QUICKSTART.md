# Quickstart
A Taproot-based server needs to do a few simple things to configure and start the application server. Specifically:

- Set up your site structure if you're serving files. A sample structure (including some representative filenames) is:
~~~
/site
  /db
    my-sqlite-database.db
  /policies
    /admin
      admin-deny-public.acacia
      admin-read-only.acacia
      admin-root.acacia
  /views
    /components
      user-table.jsml
      login-form.jsml
      navbar.jsml
    /includes
      header.jsml
      footer.jsml
    /pages
      index.jsml
      about-us.jsml
      login.jsml
      /admin
        dashboard.jsml
        add-user.jsml
        edit-user.jsml
        user-list.jsml
  /scripts
    /views
      (Compiled JSML files go here)
    /apis
      /admin
        add-user-api.js
        edit-user-api.js
        delete-user-api.js
  /static
    app.css
    logo.png
    app.js   
~~~
- Create a config struct or point Taproot at the right directories for YAML-based configs.
  - Call `taproot.NewWithConfig()` if you've created your own config, or `taproot.New()` to have Taproot find the config files and bootstrap itself.
- Open any databases and add them to `appsvr.DBs[MyDBName] = myDB` (we assume your Taproot instance is called `appsvr`).
- Add global middleware with `appsvr.AddMiddleware()`.
- Set up endpoints with `appsvr.Handler()`.
- Call `appsvr.ListenAndServeTLS()` or any other of the start functions.