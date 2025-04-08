# Goji

---

Goji is a content management platform written in Go with the intent of
replicating the ease of installation and setup of many PHP-based CMS systems
but written in a modern, performant language.

## Services
Features in Goji are "services" - generally isolated chunks of 
functionality that utilize the extension system of Goji to add 
functionality to both the admin panel as well as the front end.

Services are responsible for adding themselves to administration
panels, for exposing methods to public templates, and databases.
That said, tools are provided by Goji to simplify the process
of creating tables, linking them to existing tables, so on.

There are services that are required for Goji that are loaded automatically:

* admin - The Admin service handles the admin panel including rendering services
          that expose administrative tools.
* auth - The auth service handles authentication and user session management.
* core - Core handles the root API route and web root.

Additional services are available in the `contrib` module:

* docs - Adds support for documents, which can be seen as identical to 
         WordPress posts.

:)