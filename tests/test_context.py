from zorro_core.context.provider.cgwire import CgwireProvider


def test_cgwire_provider(mocker):
    dummy_project = {
        "name": "DummyProject",
        "id": "dummy_id",
        "data": {"config": {"fps": 25}},
    }

    dummy_episode = {
        "name": "DummyEpisode",
        "id": "dummy_id",
        "data": {},
    }

    mocker.patch("gazu.set_host", new=lambda *_: None)
    mocker.patch("gazu.log_in", new=lambda *_: None)
    mocker.patch("gazu.project.get_project_by_name", return_value=dummy_project)
    mocker.patch("gazu.shot.get_episode_by_name", return_value=dummy_episode)
    mocker.patch("gazu.shot.all_sequences_for_episode", return_value=[])
    mocker.patch("gazu.shot.all_shots_for_sequence", return_value=[])
    mocker.patch("gazu.shot.all_ranges_for_shot", return_value=[])

    provider = CgwireProvider(
        "USGS",
        "EP1180",
        [],
    )

    project = provider.get_project()
    assert project is not None
    assert project.name == dummy_project["name"]

    episodes = provider.get_episodes(project)
    assert episodes is not None
    assert len(episodes) == 1
    assert episodes[0].name == dummy_episode["name"]

    sequences = provider.get_sequences(episodes[0])
    assert sequences is not None
    assert len(sequences) == 0
