/*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/
import {NgModule} from "@angular/core";
import {RouterModule, Routes} from "@angular/router";

import {LoginComponent} from "./login/login.component";
import {AuthenticatedGuard} from "./guards/authenticated-guard.service";

const routes: Routes = [
	{component: LoginComponent, path: "login"},
	{
		canLoad: [AuthenticatedGuard],
		children: [{
			loadChildren: async (): Promise<object> => import("./core/core.module")
				.then(mod => mod.CoreModule),
			path: ""
		}],
		path: "core",
	},
	{path: "", pathMatch: "full", redirectTo: "login"}
];

/**
 * AppRoutingModule provides routing configuration for the app.
 */
@NgModule({
	exports: [RouterModule],
	imports: [RouterModule.forRoot(routes, {
		initialNavigation: "enabled",
		relativeLinkResolution: "legacy"
	})],
})
export class AppRoutingModule { }
