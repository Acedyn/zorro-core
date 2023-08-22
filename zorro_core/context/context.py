import os
from pydantic import BaseModel, Field

from zorro_core.context.plugin import Plugin


class Context(BaseModel):
    plugins: list[Plugin] = Field(default_factory=list)

    def build_environment(self):
        """
        Build the key/values for a environment with the loaded plugins
        """

        environment: dict[str, str] = {}
        for plugin in self.plugins:
            for env_key, env_values in plugin.env.items():
                # Value to optionaly set
                if env_values.set is not None:
                    environment[env_key] = env_values.set

                # Paths to append
                environment[env_key] = os.path.pathsep.join(
                     environment.get(env_key, "").strip(os.path.pathsep).split(os.path.pathsep) + [path.as_posix() for path in env_values.append]
                )

                # Paths to prepend
                environment[env_key] = os.path.pathsep.join(
                        [path.as_posix() for path in env_values.prepend][::-1] + environment.get(env_key, "").strip(os.path.pathsep).split(os.path.pathsep)
                )

        return environment
