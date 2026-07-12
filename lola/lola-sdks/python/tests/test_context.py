from lola_os.context import Overrides, get_current_overrides, override


def test_no_overrides_by_default():
    assert get_current_overrides() == Overrides()


def test_override_sets_and_restores_context():
    with override(chain="polygon", budget_max_usd=5.0):
        current = get_current_overrides()
        assert current.chain == "polygon"
        assert current.budget_max_usd == 5.0
    # Restored after the block exits.
    assert get_current_overrides() == Overrides()


def test_nested_override_composes_with_outer():
    with override(chain="ethereum", budget_max_usd=10.0):
        with override(chain="polygon"):
            current = get_current_overrides()
            # Inner override wins for `chain`, outer value persists for
            # fields the inner override didn't touch.
            assert current.chain == "polygon"
            assert current.budget_max_usd == 10.0
        # Back to just the outer override.
        assert get_current_overrides().chain == "ethereum"


def test_to_rpc_params_only_includes_set_fields():
    o = Overrides(chain="solana")
    params = o.to_rpc_params()
    assert params == {"chain": "solana"}


def test_to_rpc_params_empty_when_nothing_set():
    assert Overrides().to_rpc_params() == {}
