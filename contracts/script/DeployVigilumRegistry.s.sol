// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Script, console} from "forge-std/Script.sol";
import {VigilumRegistry} from "../src/VigilumRegistry.sol";

/**
 * @title DeployVigilumRegistry
 * @notice Deploys VigilumRegistry contract to testnet/mainnet
 * @dev Usage:
 *   forge script script/DeployVigilumRegistry.s.sol:DeployVigilumRegistry \
 *     --rpc-url $RPC_URL \
 *     --private-key $PRIVATE_KEY \
 *     --broadcast \
 *     --verify
 */
contract DeployVigilumRegistry is Script {
    function run() external returns (VigilumRegistry) {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address deployer = vm.addr(deployerPrivateKey);

        console.log("========================================");
        console.log("Deploying VigilumRegistry");
        console.log("========================================");
        console.log("Deployer:", deployer);
        console.log("Chain ID:", block.chainid);
        console.log("========================================");

        vm.startBroadcast(deployerPrivateKey);

        // Deploy VigilumRegistry
        VigilumRegistry registry = new VigilumRegistry();

        console.log("VigilumRegistry deployed at:", address(registry));
        console.log("Owner:", registry.owner());
        console.log("Deployer is authorized oracle:", registry.authorizedOracles(deployer));

        vm.stopBroadcast();

        console.log("========================================");
        console.log("Deployment Summary:");
        console.log("VigilumRegistry:", address(registry));
        console.log("========================================");
        console.log("Next steps:");
        console.log("1. Save contract address to backend config");
        console.log("2. Verify contract on Etherscan (if --verify was used)");
        console.log("3. Update ETH_CONTRACT_ADDRESS env variable");
        console.log("========================================");

        return registry;
    }
}
